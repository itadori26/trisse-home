package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"trisse-home/internal/platform/config"
)

// fakeIDP is a minimal OpenID provider that signs RS256 ID tokens.
func fakeIDP(t *testing.T, email string, appURL *string) *httptest.Server {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	b64 := base64.RawURLEncoding.EncodeToString
	var srv *httptest.Server
	nonces := map[string]string{} // code -> nonce
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": srv.URL, "authorization_endpoint": srv.URL + "/auth",
			"token_endpoint": srv.URL + "/token", "jwks_uri": srv.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{
			{"kty": "RSA", "kid": "k1", "n": b64(key.N.Bytes()), "e": b64(big.NewInt(int64(key.E)).Bytes())}}})
	})
	// The "login page" of the provider immediately redirects back with a code.
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		nonces["code-1"] = q.Get("nonce")
		http.Redirect(w, r, *appURL+"/api/auth/oidc/callback?code=code-1&state="+url.QueryEscape(q.Get("state")), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		h, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
		c, _ := json.Marshal(map[string]any{"iss": srv.URL, "sub": "user-1", "aud": "trisse", "exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(), "nonce": nonces[r.Form.Get("code")], "email": email, "email_verified": true})
		input := b64(h) + "." + b64(c)
		digest := sha256.Sum256([]byte(input))
		sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": input + "." + b64(sig)})
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestOIDCLoginEndToEnd(t *testing.T) {
	for _, tc := range []struct {
		email    string
		wantAuth bool
	}{{"me@example.org", true}, {"intruder@example.org", false}} {
		var appURL string
		idp := fakeIDP(t, tc.email, &appURL)
		app := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) {
			c.OIDCIssuer, c.OIDCClientID, c.OIDCClientSecret = idp.URL, "trisse", "secret"
			c.OIDCAllowedEmails = []string{"me@example.org"}
			c.OIDCRedirectURL = "http://placeholder/api/auth/oidc/callback"
			c.PasswordLogin = false
		})
		appURL = app.URL
		cl := newClient(t, app.URL)
		cl.do(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK)
		// A fresh browser: the password login is disabled, SSO is offered.
		b := newClient(t, app.URL)
		b.do(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123"}, http.StatusForbidden)
		var st obj
		b.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
		if sso := st["sso"].(map[string]any); sso["oidc"] != true || sso["localLogin"] != false {
			t.Fatalf("status: %v", st)
		}
		// Follow the redirects like a browser; the redirect URI points back to the app.
		b.http.CheckRedirect = nil
		res, err := b.http.Get(app.URL + "/api/auth/oidc/start")
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		b.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
		if (st["authenticated"] == true) != tc.wantAuth {
			t.Fatalf("%s: authenticated=%v (final URL %s)", tc.email, st["authenticated"], res.Request.URL)
		}
	}
}
