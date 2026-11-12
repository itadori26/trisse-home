package usecase

import (
	"context"

	"trisse-home/internal/modules/auth/domain"
)

// DemoLogin starts a session without a password. It is only wired when the
// instance runs in demo mode.
type DemoLogin interface {
	Execute(ctx context.Context, req DemoLoginRequest) (SessionResponse, error)
}

// DemoLoginRequest is the input of DemoLogin.
type DemoLoginRequest struct {
	ClientKey string
	UserAgent string
}

type demoLogin struct{ deps Deps }

// NewDemoLogin returns the DemoLogin use case.
func NewDemoLogin(deps Deps) DemoLogin { return &demoLogin{deps: deps} }

// Execute implements DemoLogin.
func (uc *demoLogin) Execute(ctx context.Context, req DemoLoginRequest) (SessionResponse, error) {
	if !uc.deps.Limiter.Allow(req.ClientKey, uc.deps.Clock.Now()) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	_, exists, err := loadAccount(ctx, uc.deps.Accounts)
	if err != nil {
		return SessionResponse{}, err
	}
	if !exists {
		return SessionResponse{}, domain.ErrSetupRequired
	}
	return uc.deps.issuer().issue(ctx, domain.OwnerID, domain.NewClient(req.UserAgent, req.ClientKey))
}
