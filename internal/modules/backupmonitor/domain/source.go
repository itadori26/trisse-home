package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of a source.
const (
	MinMaxAge        = time.Hour
	MaxMaxAge        = 60 * 24 * time.Hour
	maxNameLength    = 60
	MaxMessageLength = 1000
)

// ErrSourceNotFound is returned for unknown sources or tokens.
var ErrSourceNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The backup source does not exist.")

// State is the condition of a backup source.
type State string

// Source states.
const (
	StateOK      State = "ok"
	StateOverdue State = "overdue"
	StateFailed  State = "failed"
	StateWaiting State = "waiting"
)

// TokenHash is the SHA-256 of a ping token; only the hash is stored.
type TokenHash string

// HashToken hashes a raw ping token.
func HashToken(raw string) TokenHash {
	sum := sha256.Sum256([]byte(raw))
	return TokenHash(hex.EncodeToString(sum[:]))
}

// Source is a backup job reporting via its ping URL.
type Source struct {
	ID            string
	Name          string
	MaxAge        time.Duration
	TokenHash     TokenHash
	CreatedAt     time.Time
	LastSuccessAt time.Time
	LastFailureAt time.Time
	LastMessage   string
}

// NewSource validates and creates a source.
func NewSource(id, name string, maxAge time.Duration, token TokenHash, now time.Time) (Source, error) {
	s := Source{ID: id, TokenHash: token, CreatedAt: now}
	if err := s.Edit(name, maxAge); err != nil {
		return Source{}, err
	}
	return s, nil
}

// Edit changes name and maximum age.
func (s *Source) Edit(name string, maxAge time.Duration) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLength {
		return kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	if maxAge < MinMaxAge || maxAge > MaxMaxAge {
		return kernel.NewValidationError("maxAgeHours", "The maximum age must be between 1 hour and 60 days.")
	}
	s.Name, s.MaxAge = name, maxAge
	return nil
}

// RecordRun stores the outcome of a backup run reported via ping.
func (s *Source) RecordRun(success bool, message string, now time.Time) {
	message = strings.TrimSpace(message)
	if len(message) > MaxMessageLength {
		message = message[:MaxMessageLength]
	}
	if success {
		s.LastSuccessAt = now
	} else {
		s.LastFailureAt = now
	}
	s.LastMessage = message
}

// StateAt evaluates the source at now: a failure newer than the last
// success wins; without any success the source waits until MaxAge after its
// creation and is overdue afterwards.
func (s Source) StateAt(now time.Time) State {
	switch {
	case !s.LastFailureAt.IsZero() && s.LastFailureAt.After(s.LastSuccessAt):
		return StateFailed
	case s.LastSuccessAt.IsZero() && now.Sub(s.CreatedAt) <= s.MaxAge:
		return StateWaiting
	case s.LastSuccessAt.IsZero() || now.Sub(s.LastSuccessAt) > s.MaxAge:
		return StateOverdue
	default:
		return StateOK
	}
}
