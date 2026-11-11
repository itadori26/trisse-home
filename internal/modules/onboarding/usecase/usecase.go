package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/onboarding/domain"
	"trisse-home/internal/shared/kernel"
)

// StatusResponse describes the tour state.
type StatusResponse struct {
	Completed   bool
	CompletedAt time.Time
}

// Onboarding is the input port of the module; every method is one use case.
type Onboarding interface {
	Status(ctx context.Context) (StatusResponse, error)
	Complete(ctx context.Context) (StatusResponse, error)
	Restart(ctx context.Context) (StatusResponse, error)
}

// Repository persists the state (output port). Load returns the zero state
// when nothing has been stored yet.
type Repository interface {
	Load(ctx context.Context) (domain.State, error)
	Save(ctx context.Context, s domain.State) error
}

type onboarding struct {
	repo  Repository
	clock kernel.Clock
}

// NewOnboarding returns the onboarding use cases.
func NewOnboarding(repo Repository, clock kernel.Clock) Onboarding {
	return &onboarding{repo: repo, clock: clock}
}

func toResponse(s domain.State) StatusResponse {
	return StatusResponse{Completed: s.Completed(), CompletedAt: s.CompletedAt()}
}

// Status implements Onboarding.
func (uc *onboarding) Status(ctx context.Context) (StatusResponse, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("load onboarding state: %w", err)
	}
	return toResponse(s), nil
}

// Complete implements Onboarding.
func (uc *onboarding) Complete(ctx context.Context) (StatusResponse, error) {
	return uc.update(ctx, func(s *domain.State) { s.Complete(uc.clock.Now()) })
}

// Restart implements Onboarding.
func (uc *onboarding) Restart(ctx context.Context) (StatusResponse, error) {
	return uc.update(ctx, func(s *domain.State) { s.Restart() })
}

func (uc *onboarding) update(ctx context.Context, change func(*domain.State)) (StatusResponse, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("load onboarding state: %w", err)
	}
	change(&s)
	if err := uc.repo.Save(ctx, s); err != nil {
		return StatusResponse{}, fmt.Errorf("save onboarding state: %w", err)
	}
	return toResponse(s), nil
}
