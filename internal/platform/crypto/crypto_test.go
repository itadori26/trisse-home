package crypto

import (
	"bytes"
	"errors"
	"testing"
)

var fastParams = Argon2idParams{MemoryKiB: 1024, Iterations: 1, Parallelism: 1, SaltBytes: 16, KeyBytes: 32}

func TestArgon2idRoundTrip(t *testing.T) {
	h := NewArgon2id(fastParams)
	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := h.Verify("correct horse battery staple", hash); err != nil || !ok {
		t.Fatalf("verify correct password: ok=%v err=%v", ok, err)
	}
	if ok, _ := h.Verify("wrong", hash); ok {
		t.Fatal("wrong password verified")
	}
	if _, err := h.Verify("x", "$bcrypt$foo"); !errors.Is(err, ErrMalformedHash) {
		t.Fatalf("expected ErrMalformedHash, got %v", err)
	}
}

func TestAESGCMRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	c, err := NewAESGCM(key)
	if err != nil {
		t.Fatal(err)
	}
	enc, err := c.Encrypt("s3cret")
	if err != nil {
		t.Fatal(err)
	}
	if dec, err := c.Decrypt(enc); err != nil || dec != "s3cret" {
		t.Fatalf("decrypt = %q, %v", dec, err)
	}
	other, _ := NewAESGCM(bytes.Repeat([]byte{8}, 32))
	if _, err := other.Decrypt(enc); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("expected ErrDecrypt with wrong key, got %v", err)
	}
}

func TestRandomTokensAreUnique(t *testing.T) {
	a, _ := RandomTokens{}.NewToken()
	b, _ := RandomTokens{}.NewToken()
	if a == b || len(a) < 40 {
		t.Fatalf("weak tokens %q %q", a, b)
	}
}
