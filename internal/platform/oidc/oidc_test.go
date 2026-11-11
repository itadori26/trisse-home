package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type idp struct {
	srv      *httptest.Server
	rsaKey   *rsa.PrivateKey
	ecKey    *ecdsa.PrivateKey
	verifier string
	claims   map[string]any
	alg      string
}

var now = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func newIDP(t *testing.T) *idp {
	t.Helper()
	rk, _ := rsa.GenerateKey(rand.Reader, 2048)
	ek, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	p := &idp{rsaKey: rk, ecKey: ek, alg: "RS256"}
	mux := http.NewServeMux()
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": p.srv.URL, "authorization_endpoint": p.srv.URL + "/auth",
			"token_endpoint": p.srv.URL + "/token", "jwks_uri": p.srv.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		e := big.NewInt(int64(rk.E)).Bytes()
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{
			{"kty": "RSA", "kid": "r1", "use": "sig", "n": b64(rk.N.Bytes()), "e": b64(e)},
			{"kty": "EC", "kid": "e1", "crv": "P-256", "x": b64(ek.X.FillBytes(make([]byte, 32))), "y": b64(ek.Y.FillBytes(make([]byte, 32)))},
		}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if user, pass, _ := r.BasicAuth(); user != "client" || pass != "secret" || r.Form.Get("code") != "the-code" ||
			PKCEChallenge(r.Form.Get("code_verifier")) != PKCEChallenge(p.verifier) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": p.sign(t, p.claims)})
	})
	return p
}

func (p *idp) sign(t *testing.T, c map[string]any) string {
	t.Helper()
	kid := "r1"
	if p.alg == "ES256" {
		kid = "e1"
	}
	h, _ := json.Marshal(map[string]string{"alg": p.alg, "kid": kid})
	body, _ := json.Marshal(c)
	input := b64(h) + "." + b64(body)
	digest := sha256.Sum256([]byte(input))
	var sig []byte
	switch p.alg {
	case "RS256":
		sig, _ = rsa.SignPKCS1v15(rand.Reader, p.rsaKey, crypto.SHA256, digest[:])
	case "ES256":
		r, s, _ := ecdsa.Sign(rand.Reader, p.ecKey, digest[:])
		sig = append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...)
	case "none":
		return input + "."
	}
	return input + "." + b64(sig)
}

func (p *idp) goodClaims() map[string]any {
	return map[string]any{"iss": p.srv.URL, "sub": "user-1", "aud": "client", "exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(), "nonce": "n1", "email": "me@example.org", "email_verified": true}
}

func provider(p *idp) *Provider {
	return New(Config{Issuer: p.srv.URL + "/", ClientID: "client", ClientSecret: "secret", RedirectURL: "https://home.example/cb"},
		p.srv.Client(), func() time.Time { return now })
}

func TestAuthorizationCodeFlow(t *testing.T) {
	p := newIDP(t)
	rp := provider(p)
	p.verifier = "verifier-123"
	u, err := rp.AuthCodeURL(context.Background(), "s1", "n1", p.verifier)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.ParseQuery(strings.SplitN(u, "?", 2)[1])
	if q.Get("state") != "s1" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") != PKCEChallenge("verifier-123") {
		t.Fatalf("auth url: %s", u)
	}
	for _, alg := range []string{"RS256", "ES256"} {
		p.alg, p.claims = alg, p.goodClaims()
		id, err := rp.Exchange(context.Background(), "the-code", "verifier-123", "n1")
		if err != nil || id.Subject != "user-1" || id.Email != "me@example.org" {
			t.Fatalf("%s: %+v %v", alg, id, err)
		}
	}
	if _, err := rp.Exchange(context.Background(), "the-code", "wrong-verifier", "n1"); err == nil {
		t.Fatal("wrong PKCE verifier accepted")
	}
}

func TestVerifyRejects(t *testing.T) {
	p := newIDP(t)
	rp := provider(p)
	mutate := map[string]func(c map[string]any){
		"wrong issuer":   func(c map[string]any) { c["iss"] = "https://evil.example" },
		"wrong audience": func(c map[string]any) { c["aud"] = "other" },
		"expired":        func(c map[string]any) { c["exp"] = now.Add(-time.Hour).Unix() },
		"future iat":     func(c map[string]any) { c["iat"] = now.Add(time.Hour).Unix() },
		"wrong nonce":    func(c map[string]any) { c["nonce"] = "n2" },
		"no subject":     func(c map[string]any) { c["sub"] = "" },
	}
	for name, m := range mutate {
		c := p.goodClaims()
		m(c)
		if _, err := rp.Verify(context.Background(), p.sign(t, c), "n1"); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := rp.Verify(context.Background(), p.sign(t, p.goodClaims()), "n1"); err != nil {
		t.Fatalf("good token rejected: %v", err)
	}
	p.alg = "none"
	if _, err := rp.Verify(context.Background(), p.sign(t, p.goodClaims()), "n1"); err == nil {
		t.Fatal("alg none accepted")
	}
	p.alg = "RS256"
	good := p.sign(t, p.goodClaims())
	parts := strings.Split(good, ".")
	tampered := p.goodClaims()
	tampered["sub"] = "admin"
	body, _ := json.Marshal(tampered)
	if _, err := rp.Verify(context.Background(), parts[0]+"."+b64(body)+"."+parts[2], "n1"); err == nil {
		t.Fatal("tampered claims accepted")
	}
	audList := p.goodClaims()
	audList["aud"] = []string{"x", "client"}
	if _, err := rp.Verify(context.Background(), p.sign(t, audList), "n1"); err != nil {
		t.Fatalf("audience list rejected: %v", err)
	}
}

func TestDiscoveryIssuerMismatch(t *testing.T) {
	p := newIDP(t)
	rp := New(Config{Issuer: "https://other.example", ClientID: "client"}, p.srv.Client(), func() time.Time { return now })
	rp.cfg.Issuer = p.srv.URL + "/x"
	if _, err := rp.AuthCodeURL(context.Background(), "s", "n", "v"); err == nil {
		t.Fatal("issuer mismatch accepted")
	}
}
