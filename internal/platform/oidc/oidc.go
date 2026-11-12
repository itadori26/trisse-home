// Package oidc is a small OpenID Connect relying party: discovery, the
// authorization code flow with PKCE (S256) and ID token validation (RS256,
// ES256) against the provider's JWKS. It implements only what trisse-home
// needs and uses the standard library only (ADR 0013).
//
// Ring: Frameworks & Drivers.
package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// clockSkew tolerates small clock differences between provider and server.
const clockSkew = 2 * time.Minute

// jwksTTL bounds how long signing keys are cached (key rotation).
const jwksTTL = time.Hour

// Config identifies the client at the provider.
type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Identity is the verified subject of an ID token.
type Identity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// Provider talks to one OpenID provider.
type Provider struct {
	cfg    Config
	client *http.Client
	now    func() time.Time

	mu        sync.Mutex
	discovery *discovery
	keys      map[string]crypto.PublicKey
	keysAt    time.Time
}

type discovery struct {
	Issuer        string `json:"issuer"`
	AuthEndpoint  string `json:"authorization_endpoint"`
	TokenEndpoint string `json:"token_endpoint"`
	JWKSURI       string `json:"jwks_uri"`
}

// New returns a provider; discovery happens lazily on first use.
func New(cfg Config, client *http.Client, now func() time.Time) *Provider {
	cfg.Issuer = strings.TrimRight(cfg.Issuer, "/")
	return &Provider{cfg: cfg, client: client, now: now}
}

func (p *Provider) getJSON(ctx context.Context, u string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", u, res.Status)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(v)
}

func (p *Provider) discover(ctx context.Context) (*discovery, error) {
	p.mu.Lock()
	d := p.discovery
	p.mu.Unlock()
	if d != nil {
		return d, nil
	}
	var doc discovery
	if err := p.getJSON(ctx, p.cfg.Issuer+"/.well-known/openid-configuration", &doc); err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	if strings.TrimRight(doc.Issuer, "/") != p.cfg.Issuer {
		return nil, fmt.Errorf("oidc discovery: issuer %q does not match %q", doc.Issuer, p.cfg.Issuer)
	}
	if doc.AuthEndpoint == "" || doc.TokenEndpoint == "" || doc.JWKSURI == "" {
		return nil, errors.New("oidc discovery: endpoints missing")
	}
	p.mu.Lock()
	p.discovery = &doc
	p.mu.Unlock()
	return &doc, nil
}

// PKCEChallenge returns the S256 code challenge for verifier.
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// AuthCodeURL returns the URL the browser is sent to.
func (p *Provider) AuthCodeURL(ctx context.Context, state, nonce, verifier string) (string, error) {
	d, err := p.discover(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {p.cfg.ClientID},
		"redirect_uri":          {p.cfg.RedirectURL},
		"scope":                 {"openid email profile"},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {PKCEChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}
	sep := "?"
	if strings.Contains(d.AuthEndpoint, "?") {
		sep = "&"
	}
	return d.AuthEndpoint + sep + q.Encode(), nil
}

// Exchange redeems the authorization code and returns the verified identity.
func (p *Provider) Exchange(ctx context.Context, code, verifier, nonce string) (Identity, error) {
	d, err := p.discover(ctx)
	if err != nil {
		return Identity{}, err
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {p.cfg.RedirectURL},
		"code_verifier": {verifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(p.cfg.ClientID), url.QueryEscape(p.cfg.ClientSecret))
	res, err := p.client.Do(req)
	if err != nil {
		return Identity{}, fmt.Errorf("oidc token request: %w", err)
	}
	defer res.Body.Close()
	var body struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return Identity{}, fmt.Errorf("oidc token response: %w", err)
	}
	if res.StatusCode != http.StatusOK || body.IDToken == "" {
		return Identity{}, fmt.Errorf("oidc token request failed: %s %s", res.Status, body.Error)
	}
	return p.Verify(ctx, body.IDToken, nonce)
}

type header struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type claims struct {
	Iss           string          `json:"iss"`
	Sub           string          `json:"sub"`
	Aud           json.RawMessage `json:"aud"`
	Exp           int64           `json:"exp"`
	Iat           int64           `json:"iat"`
	Nonce         string          `json:"nonce"`
	Email         string          `json:"email"`
	EmailVerified bool            `json:"email_verified"`
	Name          string          `json:"name"`
}

func (c claims) audience() []string {
	var one string
	if json.Unmarshal(c.Aud, &one) == nil {
		return []string{one}
	}
	var many []string
	_ = json.Unmarshal(c.Aud, &many)
	return many
}

// Verify validates an ID token: signature (RS256/ES256 only), issuer,
// audience, expiry, issue time and nonce.
func (p *Provider) Verify(ctx context.Context, raw, nonce string) (Identity, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return Identity{}, errors.New("id token: malformed")
	}
	var h header
	if err := decodePart(parts[0], &h); err != nil {
		return Identity{}, fmt.Errorf("id token header: %w", err)
	}
	if h.Alg != "RS256" && h.Alg != "ES256" {
		return Identity{}, fmt.Errorf("id token: algorithm %q not allowed", h.Alg)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Identity{}, errors.New("id token: bad signature encoding")
	}
	key, err := p.key(ctx, h.Kid)
	if err != nil {
		return Identity{}, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := verifySignature(h.Alg, key, digest[:], sig); err != nil {
		return Identity{}, err
	}
	var c claims
	if err := decodePart(parts[1], &c); err != nil {
		return Identity{}, fmt.Errorf("id token claims: %w", err)
	}
	now := p.now()
	switch {
	case strings.TrimRight(c.Iss, "/") != p.cfg.Issuer:
		return Identity{}, errors.New("id token: wrong issuer")
	case !contains(c.audience(), p.cfg.ClientID):
		return Identity{}, errors.New("id token: wrong audience")
	case c.Exp == 0 || now.After(time.Unix(c.Exp, 0).Add(clockSkew)):
		return Identity{}, errors.New("id token: expired")
	case c.Iat != 0 && time.Unix(c.Iat, 0).After(now.Add(clockSkew)):
		return Identity{}, errors.New("id token: issued in the future")
	case nonce == "" || c.Nonce != nonce:
		return Identity{}, errors.New("id token: wrong nonce")
	case c.Sub == "":
		return Identity{}, errors.New("id token: no subject")
	}
	return Identity{Subject: c.Sub, Email: c.Email, EmailVerified: c.EmailVerified, Name: c.Name}, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func decodePart(s string, v any) error {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func verifySignature(alg string, key crypto.PublicKey, digest, sig []byte) error {
	switch alg {
	case "RS256":
		k, ok := key.(*rsa.PublicKey)
		if !ok || rsa.VerifyPKCS1v15(k, crypto.SHA256, digest, sig) != nil {
			return errors.New("id token: invalid signature")
		}
	case "ES256":
		k, ok := key.(*ecdsa.PublicKey)
		if !ok || len(sig) != 64 {
			return errors.New("id token: invalid signature")
		}
		r, s := new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])
		if !ecdsa.Verify(k, digest, r, s) {
			return errors.New("id token: invalid signature")
		}
	}
	return nil
}

// key returns the signing key kid, refreshing the JWKS when it is unknown
// or older than jwksTTL (key rotation).
func (p *Provider) key(ctx context.Context, kid string) (crypto.PublicKey, error) {
	p.mu.Lock()
	k, ok := p.keys[kid]
	fresh := p.now().Sub(p.keysAt) < jwksTTL
	p.mu.Unlock()
	if ok && fresh {
		return k, nil
	}
	d, err := p.discover(ctx)
	if err != nil {
		return nil, err
	}
	var set struct {
		Keys []jwk `json:"keys"`
	}
	if err := p.getJSON(ctx, d.JWKSURI, &set); err != nil {
		return nil, fmt.Errorf("oidc jwks: %w", err)
	}
	keys := map[string]crypto.PublicKey{}
	for _, j := range set.Keys {
		if pk, err := j.publicKey(); err == nil {
			keys[j.Kid] = pk
		}
	}
	p.mu.Lock()
	p.keys, p.keysAt = keys, p.now()
	p.mu.Unlock()
	if k, ok := keys[kid]; ok {
		return k, nil
	}
	return nil, errors.New("id token: unknown signing key")
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func (j jwk) publicKey() (crypto.PublicKey, error) {
	if j.Use != "" && j.Use != "sig" {
		return nil, errors.New("not a signing key")
	}
	switch j.Kty {
	case "RSA":
		n, err1 := base64.RawURLEncoding.DecodeString(j.N)
		e, err2 := base64.RawURLEncoding.DecodeString(j.E)
		if err1 != nil || err2 != nil || len(e) > 4 {
			return nil, errors.New("bad rsa key")
		}
		exp := 0
		for _, b := range e {
			exp = exp<<8 | int(b)
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exp}, nil
	case "EC":
		if j.Crv != "P-256" {
			return nil, errors.New("unsupported curve")
		}
		x, err1 := base64.RawURLEncoding.DecodeString(j.X)
		y, err2 := base64.RawURLEncoding.DecodeString(j.Y)
		if err1 != nil || err2 != nil || len(x) > 32 || len(y) > 32 {
			return nil, errors.New("bad ec key")
		}
		point := make([]byte, 65)
		point[0] = 4 // uncompressed
		copy(point[33-len(x):33], x)
		copy(point[65-len(y):], y)
		// Rejects points that are not on the curve.
		return ecdsa.ParseUncompressedPublicKey(elliptic.P256(), point)
	}
	return nil, errors.New("unsupported key type")
}
