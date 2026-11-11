package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// Clock is the kernel clock, re-declared for readable constructor signatures.
type Clock = kernel.Clock

// SessionTTL is the configured lifetime of a session.
type SessionTTL time.Duration

// Duration returns the TTL as time.Duration.
func (t SessionTTL) Duration() time.Duration { return time.Duration(t) }

// sessionIssuer creates sessions; shared by setup, login and password change.
type sessionIssuer struct {
	sessions SessionRepository
	tokens   TokenGenerator
	ids      kernel.IDGenerator
	clock    kernel.Clock
	ttl      time.Duration
}

func (i sessionIssuer) issue(ctx context.Context, client domain.Client) (SessionResponse, error) {
	raw, err := i.tokens.NewToken()
	if err != nil {
		return SessionResponse{}, fmt.Errorf("generate session token: %w", err)
	}
	csrf, err := i.tokens.NewToken()
	if err != nil {
		return SessionResponse{}, fmt.Errorf("generate csrf token: %w", err)
	}
	s, err := domain.NewSession(domain.HashToken(raw), i.ids.NewID(), csrf, client, i.clock.Now(), i.ttl)
	if err != nil {
		return SessionResponse{}, err
	}
	if err := i.sessions.Create(ctx, s); err != nil {
		return SessionResponse{}, fmt.Errorf("store session: %w", err)
	}
	return SessionResponse{SessionToken: raw, CSRFToken: csrf, ExpiresAt: s.ExpiresAt()}, nil
}

// validSession loads the session for raw and rejects unknown or expired ones.
func validSession(ctx context.Context, repo SessionRepository, raw string, now time.Time) (domain.Session, error) {
	if raw == "" {
		return domain.Session{}, domain.ErrSessionInvalid
	}
	s, err := repo.Find(ctx, domain.HashToken(raw))
	if errors.Is(err, domain.ErrSessionNotFound) {
		return domain.Session{}, domain.ErrSessionInvalid
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("find session: %w", err)
	}
	if s.IsExpired(now) {
		return domain.Session{}, domain.ErrSessionInvalid
	}
	return s, nil
}

// loadAccount returns the account and whether it exists.
func loadAccount(ctx context.Context, accounts AccountRepository) (domain.Account, bool, error) {
	a, err := accounts.Get(ctx)
	if errors.Is(err, domain.ErrAccountNotFound) {
		return domain.Account{}, false, nil
	}
	if err != nil {
		return domain.Account{}, false, fmt.Errorf("load account: %w", err)
	}
	return a, true, nil
}
