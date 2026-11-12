// Package webauthn verifies WebAuthn registrations and assertions for
// passkeys used as a second factor. It is a deliberately small relying
// party implementation on the standard library (ADR 0014): attestation
// statements are not evaluated (the server requests attestation "none"),
// and ES256, RS256 and Ed25519 credentials are supported.
//
// Ring: Frameworks & Drivers.
package webauthn

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

// COSE algorithm identifiers.
const (
	AlgES256 = -7
	AlgEdDSA = -8
	AlgRS256 = -257
)

// Authenticator data flags.
const (
	flagUserPresent  = 0x01
	flagUserVerified = 0x04
	flagAttested     = 0x40
)

// RelyingParty identifies this server to authenticators.
type RelyingParty struct {
	// ID is the host name, e.g. "home.example.org".
	ID string
	// Origin is the exact origin of the web app, e.g. "https://home.example.org".
	Origin string
}

// Credential is a registered public key.
type Credential struct {
	ID        []byte
	PublicKey []byte // PKIX DER
	Alg       int
	SignCount uint32
}

// ErrVerification is returned for every rejected registration or assertion.
var ErrVerification = errors.New("webauthn verification failed")

func fail(reason string) error { return fmt.Errorf("%w: %s", ErrVerification, reason) }

type clientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

func (rp RelyingParty) checkClientData(raw []byte, typ string, challenge []byte) error {
	var c clientData
	if err := json.Unmarshal(raw, &c); err != nil {
		return fail("client data is not JSON")
	}
	if c.Type != typ {
		return fail("wrong ceremony type")
	}
	got, err := base64.RawURLEncoding.DecodeString(c.Challenge)
	if err != nil || len(challenge) == 0 || !bytes.Equal(got, challenge) {
		return fail("challenge mismatch")
	}
	if c.Origin != rp.Origin {
		return fail("origin mismatch")
	}
	return nil
}

type authData struct {
	rpIDHash  []byte
	flags     byte
	signCount uint32
	credID    []byte
	coseKey   map[any]any
}

func parseAuthData(b []byte, attested bool) (authData, error) {
	if len(b) < 37 {
		return authData{}, fail("authenticator data too short")
	}
	a := authData{rpIDHash: b[:32], flags: b[32], signCount: binary.BigEndian.Uint32(b[33:37])}
	if !attested {
		return a, nil
	}
	if a.flags&flagAttested == 0 || len(b) < 37+16+2 {
		return authData{}, fail("no attested credential data")
	}
	rest := b[37+16:]
	n := int(binary.BigEndian.Uint16(rest[:2]))
	if n == 0 || n > 1023 || len(rest) < 2+n {
		return authData{}, fail("bad credential id length")
	}
	a.credID = rest[2 : 2+n]
	key, _, err := decodeCBOR(rest[2+n:])
	if err != nil {
		return authData{}, fail("bad credential public key")
	}
	m, ok := key.(map[any]any)
	if !ok {
		return authData{}, fail("credential public key is not a map")
	}
	a.coseKey = m
	return a, nil
}

func (rp RelyingParty) checkAuthData(a authData) error {
	want := sha256.Sum256([]byte(rp.ID))
	if !bytes.Equal(a.rpIDHash, want[:]) {
		return fail("rp id mismatch")
	}
	if a.flags&flagUserPresent == 0 {
		return fail("user not present")
	}
	return nil
}

// VerifyRegistration checks a navigator.credentials.create() response.
func (rp RelyingParty) VerifyRegistration(challenge, clientDataJSON, attestationObject []byte) (Credential, error) {
	if err := rp.checkClientData(clientDataJSON, "webauthn.create", challenge); err != nil {
		return Credential{}, err
	}
	obj, _, err := decodeCBOR(attestationObject)
	if err != nil {
		return Credential{}, fail("attestation object is not CBOR")
	}
	m, ok := obj.(map[any]any)
	if !ok {
		return Credential{}, fail("attestation object is not a map")
	}
	raw, ok := m["authData"].([]byte)
	if !ok {
		return Credential{}, fail("no authenticator data")
	}
	a, err := parseAuthData(raw, true)
	if err != nil {
		return Credential{}, err
	}
	if err := rp.checkAuthData(a); err != nil {
		return Credential{}, err
	}
	pub, alg, err := publicKeyFromCOSE(a.coseKey)
	if err != nil {
		return Credential{}, err
	}
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return Credential{}, fail("cannot encode public key")
	}
	return Credential{ID: append([]byte(nil), a.credID...), PublicKey: der, Alg: alg, SignCount: a.signCount}, nil
}

// VerifyAssertion checks a navigator.credentials.get() response for cred
// and returns the new signature counter.
func (rp RelyingParty) VerifyAssertion(challenge []byte, cred Credential, clientDataJSON, authenticatorData, signature []byte) (uint32, error) {
	if err := rp.checkClientData(clientDataJSON, "webauthn.get", challenge); err != nil {
		return 0, err
	}
	a, err := parseAuthData(authenticatorData, false)
	if err != nil {
		return 0, err
	}
	if err := rp.checkAuthData(a); err != nil {
		return 0, err
	}
	pub, err := x509.ParsePKIXPublicKey(cred.PublicKey)
	if err != nil {
		return 0, fail("stored key unreadable")
	}
	cdHash := sha256.Sum256(clientDataJSON)
	signed := append(append([]byte(nil), authenticatorData...), cdHash[:]...)
	if err := verify(cred.Alg, pub, signed, signature); err != nil {
		return 0, err
	}
	return a.signCount, nil
}

func verify(alg int, pub crypto.PublicKey, signed, sig []byte) error {
	switch alg {
	case AlgES256:
		k, ok := pub.(*ecdsa.PublicKey)
		digest := sha256.Sum256(signed)
		if !ok || !ecdsa.VerifyASN1(k, digest[:], sig) {
			return fail("invalid signature")
		}
	case AlgRS256:
		k, ok := pub.(*rsa.PublicKey)
		digest := sha256.Sum256(signed)
		if !ok || rsa.VerifyPKCS1v15(k, crypto.SHA256, digest[:], sig) != nil {
			return fail("invalid signature")
		}
	case AlgEdDSA:
		k, ok := pub.(ed25519.PublicKey)
		if !ok || !ed25519.Verify(k, signed, sig) {
			return fail("invalid signature")
		}
	default:
		return fail("unsupported algorithm")
	}
	return nil
}

func bytesField(m map[any]any, k int64) ([]byte, bool) {
	b, ok := m[k].([]byte)
	return b, ok
}

// publicKeyFromCOSE converts a COSE_Key (RFC 9053) to a Go public key.
func publicKeyFromCOSE(m map[any]any) (crypto.PublicKey, int, error) {
	kty, _ := m[int64(1)].(int64)
	alg, _ := m[int64(3)].(int64)
	switch {
	case kty == 2 && alg == AlgES256:
		crv, _ := m[int64(-1)].(int64)
		x, okx := bytesField(m, -2)
		y, oky := bytesField(m, -3)
		if crv != 1 || !okx || !oky || len(x) != 32 || len(y) != 32 {
			return nil, 0, fail("bad EC2 key")
		}
		k, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), append(append([]byte{4}, x...), y...))
		if err != nil {
			return nil, 0, fail("EC point not on curve")
		}
		return k, AlgES256, nil
	case kty == 1 && alg == AlgEdDSA:
		crv, _ := m[int64(-1)].(int64)
		x, ok := bytesField(m, -2)
		if crv != 6 || !ok || len(x) != ed25519.PublicKeySize {
			return nil, 0, fail("bad OKP key")
		}
		return ed25519.PublicKey(x), AlgEdDSA, nil
	case kty == 3 && alg == AlgRS256:
		n, okn := bytesField(m, -1)
		e, oke := bytesField(m, -2)
		if !okn || !oke || len(n) < 256 || len(e) == 0 || len(e) > 4 {
			return nil, 0, fail("bad RSA key")
		}
		exp := 0
		for _, b := range e {
			exp = exp<<8 | int(b)
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exp}, AlgRS256, nil
	}
	return nil, 0, fail("unsupported key type or algorithm")
}
