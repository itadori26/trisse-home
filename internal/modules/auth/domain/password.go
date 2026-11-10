package domain

import (
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Password length limits. The upper bound keeps hashing cost bounded.
const (
	MinPasswordLength = 12
	MaxPasswordLength = 256
)

// Password is a plaintext password that satisfies the password policy.
type Password struct {
	value string
}

// NewPassword validates raw against the password policy.
func NewPassword(raw string) (Password, error) {
	n := utf8.RuneCountInString(raw)
	switch {
	case n < MinPasswordLength:
		return Password{}, kernel.NewValidationError("password", "The password must be at least 12 characters long.")
	case n > MaxPasswordLength:
		return Password{}, kernel.NewValidationError("password", "The password must be at most 256 characters long.")
	case strings.TrimSpace(raw) == "":
		return Password{}, kernel.NewValidationError("password", "The password must not consist of whitespace only.")
	}
	return Password{value: raw}, nil
}

// String returns the plaintext. Callers must only pass it to a hasher.
func (p Password) String() string { return p.value }
