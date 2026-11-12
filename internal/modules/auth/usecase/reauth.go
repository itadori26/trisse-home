package usecase

import (
	"context"
	"errors"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
)

// ConfirmIdentity re-checks the signed-in account's password and, when
// two-factor authentication is enabled, a current code. Other modules use
// it before dangerous operations such as restoring a backup (#46).
type ConfirmIdentity interface {
	Confirm(ctx context.Context, password, code, clientKey string) error
}

type confirmIdentity struct{ deps Deps }

// NewConfirmIdentity returns the use case.
func NewConfirmIdentity(deps Deps) ConfirmIdentity { return &confirmIdentity{deps: deps} }

// Confirm implements ConfirmIdentity; failures count towards the login
// rate limit and appear in the security log.
func (uc *confirmIdentity) Confirm(ctx context.Context, password, code, clientKey string) error {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(clientKey, now) {
		return domain.ErrTooManyAttempts
	}
	a, err := uc.deps.Accounts.Get(ctx, me(ctx))
	if err != nil {
		return err
	}
	ok, err := uc.deps.Hasher.Verify(password, a.PasswordHash())
	if err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		uc.deps.Events.loginFailed(ctx, a.ID(), domain.Client{}, uc.deps.Limiter.RecordFailure(clientKey, now))
		return domain.ErrInvalidCredentials
	}
	if a.TOTPEnabled() {
		if err := uc.deps.checkSecondFactor(ctx, &a, code, now); err != nil {
			if errors.Is(err, domain.ErrInvalidTOTPCode) {
				uc.deps.Events.loginFailed(ctx, a.ID(), domain.Client{}, uc.deps.Limiter.RecordFailure(clientKey, now))
			}
			return err
		}
		// Persist the used step / recovery code so it cannot be replayed.
		if err := uc.deps.Accounts.Update(ctx, a); err != nil {
			return fmt.Errorf("update account: %w", err)
		}
	}
	uc.deps.Limiter.Reset(clientKey)
	return nil
}
