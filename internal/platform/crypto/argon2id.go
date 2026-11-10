package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// maxKeyBytes bounds the decoded key length of a stored hash.
const maxKeyBytes = 1024

// ErrMalformedHash reports a stored hash that cannot be parsed.
var ErrMalformedHash = errors.New("malformed argon2id hash")

// Argon2idParams are the cost parameters. The defaults follow the OWASP
// recommendation for Argon2id (m=64 MiB, t=3, p=2).
type Argon2idParams struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltBytes   uint32
	KeyBytes    uint32
}

// DefaultArgon2idParams are suitable for interactive logins on small servers.
var DefaultArgon2idParams = Argon2idParams{MemoryKiB: 64 * 1024, Iterations: 3, Parallelism: 2, SaltBytes: 16, KeyBytes: 32}

// Argon2id hashes and verifies passwords in the PHC string format.
type Argon2id struct {
	params Argon2idParams
}

// NewArgon2id returns a hasher using params.
func NewArgon2id(params Argon2idParams) *Argon2id { return &Argon2id{params: params} }

// Hash returns the PHC-encoded Argon2id hash of password.
func (a *Argon2id) Hash(password string) (string, error) {
	salt := make([]byte, a.params.SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	p := a.params
	key := argon2.IDKey([]byte(password), salt, p.Iterations, p.MemoryKiB, p.Parallelism, p.KeyBytes)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.MemoryKiB, p.Iterations, p.Parallelism, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

// Verify reports whether password matches encoded. The parameters stored in
// the hash are used, so old hashes keep working after a parameter change.
func (a *Argon2id) Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, ErrMalformedHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, ErrMalformedHash
	}
	var p Argon2idParams
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.MemoryKiB, &p.Iterations, &p.Parallelism); err != nil {
		return false, ErrMalformedHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrMalformedHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) == 0 || len(want) > maxKeyBytes {
		return false, ErrMalformedHash
	}
	got := argon2.IDKey([]byte(password), salt, p.Iterations, p.MemoryKiB, p.Parallelism, uint32(len(want))) //nolint:gosec // bounded by maxKeyBytes above
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
