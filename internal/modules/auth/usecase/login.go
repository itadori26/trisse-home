package usecase

import (
	"context"
	"errors"
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
	client := domain.NewClient(req.UserAgent, req.ClientKey)
	if !ok {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
		return SessionResponse{}, domain.ErrInvalidCredentials
	}
	if account.TOTPEnabled() {
		if err := uc.deps.checkSecondFactor(ctx, &account, req.TOTPCode, now); err != nil {
			if errors.Is(err, domain.ErrInvalidTOTPCode) {
				uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
			}
			return SessionResponse{}, err
		}
		if err := uc.deps.Accounts.Update(ctx, account); err != nil {
			return SessionResponse{}, fmt.Errorf("update account: %w", err)
		}
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	uc.deps.Events.loginSucceeded(ctx, domain.OwnerID, client, "password")
	return uc.deps.issuer().issue(ctx, client)
}
