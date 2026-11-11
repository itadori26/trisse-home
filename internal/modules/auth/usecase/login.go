package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	owner, exists, err := loadAccount(ctx, uc.deps.Accounts)
	if err != nil {
		return SessionResponse{}, err
	}
	if !exists {
		return SessionResponse{}, domain.ErrSetupRequired
	}
	client := domain.NewClient(req.UserAgent, req.ClientKey)
	account, err := uc.account(ctx, owner, req.Username)
	if err != nil {
		return SessionResponse{}, err
	}
	// Unknown and disabled guests are checked against the owner's hash so
	// the response time does not reveal which names exist.
	ok, err := uc.deps.Hasher.Verify(req.Password, account.PasswordHash())
	if err != nil {
		return SessionResponse{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok || account.Disabled() || (req.Username != "" && account.IsOwner()) {
		uc.deps.Events.loginFailed(ctx, account.ID(), client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
		return SessionResponse{}, domain.ErrInvalidCredentials
	}
	if account.TOTPEnabled() {
		if err := uc.deps.checkSecondFactor(ctx, &account, req.TOTPCode, now); err != nil {
			if errors.Is(err, domain.ErrInvalidTOTPCode) {
				uc.deps.Events.loginFailed(ctx, account.ID(), client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
			}
			return SessionResponse{}, err
		}
		if err := uc.deps.Accounts.Update(ctx, account); err != nil {
			return SessionResponse{}, fmt.Errorf("update account: %w", err)
		}
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	uc.deps.Events.loginSucceeded(ctx, account.ID(), client, "password")
	return uc.deps.issuer().issue(ctx, account.ID(), client)
}

// account resolves the login name: empty means the owner; an unknown name
// yields the owner account so the password check fails in constant time.
func (uc *login) account(ctx context.Context, owner domain.Account, name string) (domain.Account, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return owner, nil
	}
	a, err := uc.deps.Accounts.FindByName(ctx, name)
	if errors.Is(err, domain.ErrGuestNotFound) || (err == nil && a.IsOwner()) {
		return owner, nil
	}
	return a, err
}
