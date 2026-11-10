package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
)

// Deps bundles the output ports shared by the auth use cases.
type Deps struct {
	Accounts AccountRepository
	Sessions SessionRepository
	Hasher   PasswordHasher
	Tokens   TokenGenerator
	Limiter  AttemptLimiter
	Clock    Clock
	TTL      SessionTTL
}

func (d Deps) issuer() sessionIssuer {
	return sessionIssuer{sessions: d.Sessions, tokens: d.Tokens, clock: d.Clock, ttl: d.TTL.Duration()}
}

type setupAccount struct {
	deps     Deps
	setup    SetupTokenVerifier
	observer SetupObserver
}

// NewSetupAccount returns the SetupAccount use case; observer is notified
// after the account was created.
func NewSetupAccount(deps Deps, setup SetupTokenVerifier, observer SetupObserver) SetupAccount {
	return &setupAccount{deps: deps, setup: setup, observer: observer}
}

// Execute implements SetupAccount. The setup token prevents anyone else in
// the network from claiming a freshly started instance.
func (uc *setupAccount) Execute(ctx context.Context, req SetupRequest) (SessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(req.ClientKey, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	_, exists, err := loadAccount(ctx, uc.deps.Accounts)
	if err != nil {
		return SessionResponse{}, err
	}
	if exists {
		return SessionResponse{}, domain.ErrAlreadySetUp
	}
	if !uc.setup.Verify(req.SetupToken) {
		uc.deps.Limiter.RecordFailure(req.ClientKey, now)
		return SessionResponse{}, domain.ErrInvalidSetupToken
	}
	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return SessionResponse{}, err
	}
	hash, err := uc.deps.Hasher.Hash(password.String())
	if err != nil {
		return SessionResponse{}, fmt.Errorf("hash password: %w", err)
	}
	account, err := domain.NewAccount(hash, now)
	if err != nil {
		return SessionResponse{}, err
	}
	if err := uc.deps.Accounts.Create(ctx, account); err != nil {
		return SessionResponse{}, err
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	uc.observer.AccountCreated(ctx)
	return uc.deps.issuer().issue(ctx)
}
