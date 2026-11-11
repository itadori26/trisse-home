package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
)

type login struct{ deps Deps }

// NewLogin returns the Login use case.
func NewLogin(deps Deps) Login { return &login{deps: deps} }

// Execute implements Login.
func (uc *login) Execute(ctx context.Context, req LoginRequest) (SessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(req.ClientKey, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	account, exists, err := loadAccount(ctx, uc.deps.Accounts)
	if err != nil {
		return SessionResponse{}, err
	}
	if !exists {
		return SessionResponse{}, domain.ErrSetupRequired
	}
	ok, err := uc.deps.Hasher.Verify(req.Password, account.PasswordHash())
	if err != nil {
		return SessionResponse{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		uc.deps.Limiter.RecordFailure(req.ClientKey, now)
		return SessionResponse{}, domain.ErrInvalidCredentials
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	return uc.deps.issuer().issue(ctx, domain.NewClient(req.UserAgent, req.ClientKey))
}
