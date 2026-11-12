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
	// StateRunning: a /start ping arrived and the run has not finished yet.
	StateRunning State = "running"
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
	ID   string
	Name string
	// Kind is "backup" or "check" (any heartbeat, #100).
	Kind string
	// MaxAge is the interval schedule; with Cron set it is unused.
	MaxAge time.Duration
	// Cron is an optional cron schedule; Grace is added to either schedule.
	Cron          *Cron
	Grace         time.Duration
	LastStartAt   time.Time
	LastDuration  time.Duration
	TokenHash     TokenHash
	CreatedAt     time.Time
	LastSuccessAt time.Time
	LastFailureAt time.Time
	LastMessage   string
}

// Schedule describes when a source is expected to report.
type Schedule struct {
	Kind   string // "backup" (default) or "check"
	MaxAge time.Duration
	// Cron replaces MaxAge when set, e.g. "0 3 * * *".
	Cron  string
	Grace time.Duration
}

// MaxGrace bounds the grace period.
const MaxGrace = 7 * 24 * time.Hour

// NewSource creates a source; schedule is validated by Edit.
func NewSource(id, name string, sch Schedule, token TokenHash, now time.Time) (Source, error) {
	s := Source{ID: id, TokenHash: token, CreatedAt: now}
	if err := s.Edit(name, sch); err != nil {
		return Source{}, err
	}
	return s, nil
}

// Edit changes the name and the schedule.
func (s *Source) Edit(name string, sch Schedule) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLength {
		return kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	kind := sch.Kind
	if kind == "" {
		kind = "backup"
	}
	if kind != "backup" && kind != "check" {
		return kernel.NewValidationError("kind", "The kind must be backup or check.")
	}
	if sch.Grace < 0 || sch.Grace > MaxGrace {
		return kernel.NewValidationError("graceMinutes", "The grace period must be between 0 minutes and 7 days.")
	}
	var cron *Cron
	if strings.TrimSpace(sch.Cron) != "" {
		c, err := ParseCron(sch.Cron)
		if err != nil {
			return err
		}
		cron = &c
	} else if sch.MaxAge < MinMaxAge || sch.MaxAge > MaxMaxAge {
		return kernel.NewValidationError("maxAgeHours", "The maximum age must be between 1 hour and 60 days.")
	}
	s.Name, s.Kind, s.MaxAge, s.Cron, s.Grace = name, kind, sch.MaxAge, cron, sch.Grace
	return nil
}

// Schedule returns the schedule for views and storage.
func (s Source) Schedule() Schedule {
	sch := Schedule{Kind: s.Kind, MaxAge: s.MaxAge, Grace: s.Grace}
	if s.Cron != nil {
		sch.Cron = s.Cron.String()
	}
	return sch
}

// RecordStart notes a /start ping; the next success measures the duration.
func (s *Source) RecordStart(now time.Time) { s.LastStartAt = now }

// RecordRun stores the result of a run.
func (s *Source) RecordRun(success bool, message string, now time.Time) {
	message = strings.TrimSpace(message)
	if len(message) > MaxMessageLength {
		message = message[:MaxMessageLength]
	}
	if s.running(now) {
		s.LastDuration = now.Sub(s.LastStartAt)
	}
	if success {
		s.LastSuccessAt = now
	} else {
		s.LastFailureAt = now
	}
	s.LastMessage = message
}

// maxRun bounds how long a started run counts as running.
const maxRun = 24 * time.Hour

func (s Source) running(now time.Time) bool {
	last := s.LastSuccessAt
	if s.LastFailureAt.After(last) {
		last = s.LastFailureAt
	}
	return !s.LastStartAt.IsZero() && s.LastStartAt.After(last) && now.Sub(s.LastStartAt) < maxRun
}

// StateAt evaluates the source at now.
func (s Source) StateAt(now time.Time) State {
	failed := !s.LastFailureAt.IsZero() && s.LastFailureAt.After(s.LastSuccessAt)
	switch {
	case s.running(now) && !s.overdue(now):
		return StateRunning
	case failed:
		return StateFailed
	case s.overdue(now):
		if s.LastSuccessAt.IsZero() && !s.dueSinceCreation(now) {
			return StateWaiting
		}
		return StateOverdue
	case s.LastSuccessAt.IsZero():
		return StateWaiting
	default:
		return StateOK
	}
}

// overdue reports whether the expected run plus grace has passed without
// a successful ping.
func (s Source) overdue(now time.Time) bool {
	if s.Cron != nil {
		// Exclusive: exactly at the end of the grace period it is not overdue yet.
		due := s.Cron.Prev(now.Add(-s.Grace - time.Nanosecond))
		return !due.IsZero() && due.After(s.CreatedAt) && s.LastSuccessAt.Before(due)
	}
	since := s.LastSuccessAt
	if since.IsZero() {
		since = s.CreatedAt
	}
	return now.Sub(since) > s.MaxAge+s.Grace
}

// dueSinceCreation: a never-run source only becomes overdue after its first
// expected run.
func (s Source) dueSinceCreation(now time.Time) bool {
	if s.Cron != nil {
		return true // overdue already checks that the due time is after creation
	}
	return now.Sub(s.CreatedAt) > s.MaxAge+s.Grace
}
