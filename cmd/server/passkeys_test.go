package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"trisse-home/internal/platform/config"
)

// cbor encodes the few CBOR items a "none" attestation needs.
func cborItem(major byte, n int) []byte {
	if n < 24 {
		return []byte{major<<5 | byte(n)}
	}
	if n < 256 {
		return []byte{major<<5 | 24, byte(n)}
	}
	return []byte{major<<5 | 25, byte(n >> 8), byte(n)}
}

func cborIntV(v int) []byte {
	if v >= 0 {
		return cborItem(0, v)
	}
	return cborItem(1, -1-v)
}

func TestPasskeyLoginEndToEnd(t *testing.T) {
	const origin = "http://localhost"
	srv := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) { c.PublicURL, _ = url.Parse(origin) })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pt, _ := key.PublicKey.Bytes()
	credID := []byte("e2e-credential")
	b64 := base64.RawURLEncoding.EncodeToString
	rpHash := sha256.Sum256([]byte("localhost"))
	clientData := func(typ, challenge string) []byte {
		b, _ := json.Marshal(map[string]string{"type": typ, "challenge": challenge, "origin": origin})
		return b
	}

	// Registration.
	var opts obj
	c.json(http.MethodPost, "/api/auth/passkeys/register/begin", nil, http.StatusOK, &opts)
	cose := append(cborItem(5, 5), cborIntV(1)...)
	for _, kv := range [][]byte{cborIntV(2), cborIntV(3), cborIntV(-7), cborIntV(-1), cborIntV(1),
		cborIntV(-2), append(cborItem(2, 32), pt[1:33]...), cborIntV(-3), append(cborItem(2, 32), pt[33:]...)} {
		cose = append(cose, kv...)
	}
	ad := append(append([]byte(nil), rpHash[:]...), 0x45, 0, 0, 0, 1) // UP|UV|AT, counter 1
	ad = append(ad, make([]byte, 16)...)
	ad = binary.BigEndian.AppendUint16(ad, uint16(len(credID)))
	ad = append(append(ad, credID...), cose...)
	att := append(cborItem(5, 3), append(cborItem(3, 3), "fmt"...)...)
	att = append(att, append(cborItem(3, 4), "none"...)...)
	att = append(att, append(cborItem(3, 7), "attStmt"...)...)
	att = append(att, cborItem(5, 0)...)
	att = append(att, append(cborItem(3, 8), "authData"...)...)
	att = append(att, append(cborItem(2, len(ad)), ad...)...)
	var reg obj
	c.json(http.MethodPost, "/api/auth/passkeys/register/finish", obj{"name": "Test key",
		"clientDataJSON": b64(clientData("webauthn.create", opts["challenge"].(string))), "attestationObject": b64(att)}, http.StatusCreated, &reg)
	if codes := reg["recoveryCodes"].([]any); len(codes) != 10 {
		t.Fatalf("recovery codes: %v", reg)
	}

	// Login: password, then the passkey.
	b := newClient(t, srv.URL)
	var sf struct {
		SecondFactor struct {
			Ticket, Challenge string
			AllowCredentials  []string
		}
	}
	b.json(http.MethodPost, "/api/auth/login", obj{"password": "e2e-password-123"}, http.StatusAccepted, &sf)
	if sf.SecondFactor.Ticket == "" || len(sf.SecondFactor.AllowCredentials) != 1 {
		t.Fatalf("second factor: %+v", sf)
	}
	assert := func(counter uint32) obj {
		cd := clientData("webauthn.get", sf.SecondFactor.Challenge)
		authData := binary.BigEndian.AppendUint32(append(append([]byte(nil), rpHash[:]...), 0x05), counter)
		h := sha256.Sum256(cd)
		d := sha256.Sum256(append(append([]byte(nil), authData...), h[:]...))
		sig, _ := ecdsa.SignASN1(rand.Reader, key, d[:])
		return obj{"ticket": sf.SecondFactor.Ticket, "credentialId": sf.SecondFactor.AllowCredentials[0],
			"clientDataJSON": b64(cd), "authenticatorData": b64(authData), "signature": b64(sig)}
	}
	b.do(http.MethodPost, "/api/auth/passkeys/login", assert(2), http.StatusOK)
	var st obj
	b.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
	if st["authenticated"] != true {
		t.Fatalf("not signed in: %v", st)
	}
	// The ticket is single use.
	b.do(http.MethodPost, "/api/auth/passkeys/login", assert(3), http.StatusUnauthorized)
}
