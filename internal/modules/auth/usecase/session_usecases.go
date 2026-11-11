package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
)

type logout struct{ sessions SessionRepository }

// NewLogout returns the Logout use case.
func NewLogout(sessions SessionRepository) Logout { return &logout{sessions: sessions} }

// Execute implements Logout. Logging out an unknown session is not an error.
func (uc *logout) Execute(ctx context.Context, req LogoutRequest) error {
	if req.SessionToken == "" {
		return nil
	}
	if err := uc.sessions.Delete(ctx, domain.HashToken(req.SessionToken)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

type authenticate struct {
	sessions SessionRepository
	clock    Clock
	ttl      time.Duration
}

// NewAuthenticate returns the Authenticate use case.
func NewAuthenticate(sessions SessionRepository, clock Clock, ttl SessionTTL) Authenticate {
	return &authenticate{sessions: sessions, clock: clock, ttl: ttl.Duration()}
}

// Execute implements Authenticate and extends sliding sessions.
func (uc *authenticate) Execute(ctx context.Context, req AuthenticateRequest) (AuthenticateResponse, error) {
	now := uc.clock.Now()
	s, err := validSession(ctx, uc.sessions, req.SessionToken, now)
	if err != nil {
		return AuthenticateResponse{}, err
	}
	refreshed := s.Refresh(now, uc.ttl)
	if refreshed {
		if err := uc.sessions.Update(ctx, s); err != nil {
			return AuthenticateResponse{}, fmt.Errorf("refresh session: %w", err)
		}
	}
	return AuthenticateResponse{CSRFToken: s.CSRFToken(), ExpiresAt: s.ExpiresAt(), Refreshed: refreshed}, nil
}

type changePassword struct{ deps Deps }

// NewChangePassword returns the ChangePassword use case.
func NewChangePassword(deps Deps) ChangePassword { return &changePassword{deps: deps} }

// Execute implements ChangePassword. All sessions are revoked so a possibly
// compromised session cannot outlive the password change.
func (uc *changePassword) Execute(ctx context.Context, req ChangePasswordRequest) (SessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(req.ClientKey, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	account, err := uc.deps.Accounts.Get(ctx)
	if err != nil {
		return SessionResponse{}, err
	}
	ok, err := uc.deps.Hasher.Verify(req.CurrentPassword, account.PasswordHash())
	if err != nil {
		return SessionResponse{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, domain.NewClient(req.UserAgent, req.ClientKey), uc.deps.Limiter.RecordFailure(req.ClientKey, now))
		return SessionResponse{}, domain.ErrInvalidCredentials
	}
	password, err := domain.NewPassword(req.NewPassword)
	if err != nil {
		return SessionResponse{}, err
	}
	hash, err := uc.deps.Hasher.Hash(password.String())
	if err != nil {
		return SessionResponse{}, fmt.Errorf("hash password: %w", err)
	}
	if err := account.ChangePasswordHash(hash, now); err != nil {
		return SessionResponse{}, err
	}
	if err := uc.deps.Accounts.Update(ctx, account); err != nil {
		return SessionResponse{}, fmt.Errorf("update account: %w", err)
	}
	if err := uc.deps.Sessions.DeleteAll(ctx); err != nil {
		return SessionResponse{}, fmt.Errorf("revoke sessions: %w", err)
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	client := domain.NewClient(req.UserAgent, req.ClientKey)
	uc.deps.Events.record(ctx, domain.EventPasswordChanged, domain.OwnerID, client, "")
	return uc.deps.issuer().issue(ctx, client)
}

type resetAccount struct {
	accounts AccountRepository
	sessions SessionRepository
}

// NewResetAccount returns the ResetAccount use case.
func NewResetAccount(accounts AccountRepository, sessions SessionRepository) ResetAccount {
	return &resetAccount{accounts: accounts, sessions: sessions}
}

// Execute implements ResetAccount.
func (uc *resetAccount) Execute(ctx context.Context) error {
	if err := uc.sessions.DeleteAll(ctx); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}
	if err := uc.accounts.Delete(ctx); err != nil && !errors.Is(err, domain.ErrAccountNotFound) {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

type purgeExpiredSessions struct {
	sessions SessionRepository
	clock    Clock
}

// NewPurgeExpiredSessions returns the PurgeExpiredSessions use case.
func NewPurgeExpiredSessions(sessions SessionRepository, clock Clock) PurgeExpiredSessions {
	return &purgeExpiredSessions{sessions: sessions, clock: clock}
}

// Execute implements PurgeExpiredSessions.
func (uc *purgeExpiredSessions) Execute(ctx context.Context) (int64, error) {
	n, err := uc.sessions.DeleteExpired(ctx, uc.clock.Now())
	if err != nil {
		return 0, fmt.Errorf("purge sessions: %w", err)
	}
	return n, nil
}

type manageSessions struct {
	sessions SessionRepository
	clock    Clock
	events   *Recorder
}

// NewManageSessions returns the ManageSessions use cases; events may be nil.
func NewManageSessions(sessions SessionRepository, clock Clock, events *Recorder) ManageSessions {
	return &manageSessions{sessions: sessions, clock: clock, events: events}
}

// List implements ManageSessions; the caller's session is marked as current.
func (uc *manageSessions) List(ctx context.Context, currentToken string) ([]SessionView, error) {
	list, err := uc.sessions.ListActive(ctx, uc.clock.Now())
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	current := domain.HashToken(currentToken)
	out := make([]SessionView, 0, len(list))
	for _, s := range list {
		out = append(out, SessionView{
			ID: s.PublicID(), UserAgent: s.Client().UserAgent, ClientIP: s.Client().IP, CreatedAt: s.CreatedAt(),
			LastSeenAt: s.RefreshedAt(), ExpiresAt: s.ExpiresAt(), Current: s.TokenHash() == current,
		})
	}
	return out, nil
}

// Revoke implements ManageSessions.
func (uc *manageSessions) Revoke(ctx context.Context, _ string, sessionID string) error {
	if err := uc.sessions.DeleteByPublicID(ctx, sessionID); err != nil {
		return err
	}
	uc.events.record(ctx, domain.EventSessionRevoked, domain.OwnerID, domain.Client{}, "")
	return nil
}

// RevokeOthers implements ManageSessions: every session except the caller's ends.
func (uc *manageSessions) RevokeOthers(ctx context.Context, currentToken string) (int64, error) {
	n, err := uc.sessions.DeleteAllExcept(ctx, domain.HashToken(currentToken))
	if err != nil {
		return 0, fmt.Errorf("revoke sessions: %w", err)
	}
	uc.events.record(ctx, domain.EventLogoutEverywhere, domain.OwnerID, domain.Client{}, fmt.Sprintf("%d sessions", n))
	return n, nil
}
