package webauthn

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"testing"
)

// --- a tiny CBOR encoder for test vectors ---

func cborHead(major byte, n int) []byte {
	switch {
	case n < 24:
		return []byte{major<<5 | byte(n)}
	case n < 256:
		return []byte{major<<5 | 24, byte(n)}
	default:
		return []byte{major<<5 | 25, byte(n >> 8), byte(n)}
	}
}

func cborInt(v int) []byte {
	if v >= 0 {
		return cborHead(0, v)
	}
	return cborHead(1, -1-v)
}

func cborBytes(b []byte) []byte { return append(cborHead(2, len(b)), b...) }
func cborText(s string) []byte  { return append(cborHead(3, len(s)), s...) }
func cborMap(pairs ...[]byte) []byte {
	out := cborHead(5, len(pairs)/2)
	for _, p := range pairs {
		out = append(out, p...)
	}
	return out
}

// --- simulated authenticator ---

type authenticator struct {
	alg     int
	ec      *ecdsa.PrivateKey
	ed      ed25519.PrivateKey
	credID  []byte
	counter uint32
}

func newAuthenticator(alg int) *authenticator {
	a := &authenticator{alg: alg, credID: []byte("credential-1")}
	if alg == AlgES256 {
		a.ec, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	} else {
		_, a.ed, _ = ed25519.GenerateKey(rand.Reader)
	}
	return a
}

func (a *authenticator) coseKey() []byte {
	if a.alg == AlgES256 {
		pt, _ := a.ec.PublicKey.Bytes()
		return cborMap(cborInt(1), cborInt(2), cborInt(3), cborInt(AlgES256), cborInt(-1), cborInt(1),
			cborInt(-2), cborBytes(pt[1:33]), cborInt(-3), cborBytes(pt[33:]))
	}
	return cborMap(cborInt(1), cborInt(1), cborInt(3), cborInt(AlgEdDSA), cborInt(-1), cborInt(6),
		cborInt(-2), cborBytes(a.ed.Public().(ed25519.PublicKey)))
}

func (a *authenticator) authData(rpID string, flags byte, attested bool) []byte {
	h := sha256.Sum256([]byte(rpID))
	out := append([]byte(nil), h[:]...)
	out = append(out, flags)
	out = binary.BigEndian.AppendUint32(out, a.counter)
	if attested {
		out = append(out, make([]byte, 16)...)
		out = binary.BigEndian.AppendUint16(out, uint16(len(a.credID)))
		out = append(out, a.credID...)
		out = append(out, a.coseKey()...)
	}
	return out
}

func clientDataJSON(typ string, challenge []byte, origin string) []byte {
	b, _ := json.Marshal(map[string]string{"type": typ, "challenge": base64.RawURLEncoding.EncodeToString(challenge), "origin": origin})
	return b
}

func (a *authenticator) sign(data []byte) []byte {
	if a.alg == AlgES256 {
		d := sha256.Sum256(data)
		sig, _ := ecdsa.SignASN1(rand.Reader, a.ec, d[:])
		return sig
	}
	sig, _ := a.ed.Sign(rand.Reader, data, crypto.Hash(0))
	return sig
}

var rp = RelyingParty{ID: "home.example.org", Origin: "https://home.example.org"}

func register(t *testing.T, a *authenticator) Credential {
	t.Helper()
	challenge := []byte("register-challenge")
	att := cborMap(cborText("fmt"), cborText("none"), cborText("attStmt"), cborMap(),
		cborText("authData"), cborBytes(a.authData(rp.ID, flagUserPresent|flagUserVerified|flagAttested, true)))
	cred, err := rp.VerifyRegistration(challenge, clientDataJSON("webauthn.create", challenge, rp.Origin), att)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(cred.ID, a.credID) || cred.Alg != a.alg {
		t.Fatalf("credential: %+v", cred)
	}
	return cred
}

func (a *authenticator) assert(challenge []byte, origin, rpID string, flags byte) (cd, ad, sig []byte) {
	a.counter++
	cd = clientDataJSON("webauthn.get", challenge, origin)
	ad = a.authData(rpID, flags, false)
	h := sha256.Sum256(cd)
	return cd, ad, a.sign(append(append([]byte(nil), ad...), h[:]...))
}

func TestPasskeyCeremonies(t *testing.T) {
	for _, alg := range []int{AlgES256, AlgEdDSA} {
		a := newAuthenticator(alg)
		cred := register(t, a)
		challenge := []byte("login-challenge")
		cd, ad, sig := a.assert(challenge, rp.Origin, rp.ID, flagUserPresent)
		count, err := rp.VerifyAssertion(challenge, cred, cd, ad, sig)
		if err != nil || count != a.counter {
			t.Fatalf("alg %d: %d %v", alg, count, err)
		}
	}
}

func TestAssertionRejects(t *testing.T) {
	a := newAuthenticator(AlgES256)
	cred := register(t, a)
	ch := []byte("c1")
	cases := map[string]func() ([]byte, []byte, []byte, []byte){
		"other challenge": func() ([]byte, []byte, []byte, []byte) {
			cd, ad, sig := a.assert([]byte("c2"), rp.Origin, rp.ID, flagUserPresent)
			return ch, cd, ad, sig
		},
		"phishing origin": func() ([]byte, []byte, []byte, []byte) {
			cd, ad, sig := a.assert(ch, "https://home-example.evil", rp.ID, flagUserPresent)
			return ch, cd, ad, sig
		},
		"other rp id": func() ([]byte, []byte, []byte, []byte) {
			cd, ad, sig := a.assert(ch, rp.Origin, "evil.example", flagUserPresent)
			return ch, cd, ad, sig
		},
		"user not present": func() ([]byte, []byte, []byte, []byte) {
			cd, ad, sig := a.assert(ch, rp.Origin, rp.ID, 0)
			return ch, cd, ad, sig
		},
		"tampered data": func() ([]byte, []byte, []byte, []byte) {
			cd, ad, sig := a.assert(ch, rp.Origin, rp.ID, flagUserPresent)
			ad[len(ad)-1]++
			return ch, cd, ad, sig
		},
		"wrong ceremony": func() ([]byte, []byte, []byte, []byte) {
			cd := clientDataJSON("webauthn.create", ch, rp.Origin)
			ad := a.authData(rp.ID, flagUserPresent, false)
			h := sha256.Sum256(cd)
			return ch, cd, ad, a.sign(append(append([]byte(nil), ad...), h[:]...))
		},
	}
	for name, mk := range cases {
		challenge, cd, ad, sig := mk()
		if _, err := rp.VerifyAssertion(challenge, cred, cd, ad, sig); !errors.Is(err, ErrVerification) {
			t.Errorf("%s accepted: %v", name, err)
		}
	}
	other := newAuthenticator(AlgES256)
	cd, ad, sig := other.assert(ch, rp.Origin, rp.ID, flagUserPresent)
	if _, err := rp.VerifyAssertion(ch, cred, cd, ad, sig); err == nil {
		t.Fatal("signature of another key accepted")
	}
}

func TestRegistrationRejects(t *testing.T) {
	a := newAuthenticator(AlgES256)
	ch := []byte("r1")
	att := cborMap(cborText("fmt"), cborText("none"), cborText("attStmt"), cborMap(),
		cborText("authData"), cborBytes(a.authData(rp.ID, flagUserPresent|flagAttested, true)))
	if _, err := rp.VerifyRegistration([]byte("r2"), clientDataJSON("webauthn.create", ch, rp.Origin), att); err == nil {
		t.Fatal("wrong challenge accepted")
	}
	if _, err := rp.VerifyRegistration(ch, clientDataJSON("webauthn.create", ch, rp.Origin), []byte{0xff, 0x00}); err == nil {
		t.Fatal("garbage accepted")
	}
	noAttested := cborMap(cborText("fmt"), cborText("none"), cborText("authData"), cborBytes(a.authData(rp.ID, flagUserPresent, false)))
	if _, err := rp.VerifyRegistration(ch, clientDataJSON("webauthn.create", ch, rp.Origin), noAttested); err == nil {
		t.Fatal("missing credential accepted")
	}
}

func TestCBORLimits(t *testing.T) {
	deep := bytes.Repeat([]byte{0x81}, 20) // nested arrays
	if _, _, err := decodeCBOR(append(deep, 0x00)); err == nil {
		t.Fatal("deep nesting accepted")
	}
	if _, _, err := decodeCBOR([]byte{0x5a, 0xff, 0xff, 0xff, 0xff}); err == nil {
		t.Fatal("oversized byte string accepted")
	}
}
