package usecase

import (
	"context"
	"errors"

	"trisse-home/internal/modules/auth/domain"
)

type getStatus struct {
	accounts AccountRepository
	sessions SessionRepository
	clock    Clock
}

// NewGetStatus returns the GetStatus use case.
func NewGetStatus(accounts AccountRepository, sessions SessionRepository, clock Clock) GetStatus {
	return &getStatus{accounts: accounts, sessions: sessions, clock: clock}
}

// Execute implements GetStatus.
func (uc *getStatus) Execute(ctx context.Context, req StatusRequest) (StatusResponse, error) {
	_, exists, err := loadAccount(ctx, uc.accounts)
	if err != nil {
		return StatusResponse{}, err
	}
	if !exists {
		return StatusResponse{SetupRequired: true}, nil
	}
	s, err := validSession(ctx, uc.sessions, req.SessionToken, uc.clock.Now())
	if errors.Is(err, domain.ErrSessionInvalid) {
		return StatusResponse{}, nil
	}
	if err != nil {
		return StatusResponse{}, err
	}
	a, err := uc.accounts.Get(ctx, s.AccountID())
	if err != nil || a.Disabled() {
		return StatusResponse{}, nil //nolint:nilerr // a session of a removed or disabled account is simply not signed in
	}
	return StatusResponse{Authenticated: true, CSRFToken: s.CSRFToken(), Role: string(a.Role()), Username: a.Name()}, nil
}
