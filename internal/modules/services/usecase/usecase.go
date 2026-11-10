// Package usecase contains the services use cases: managing tiles and
// checking their health.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/kernel"
)

// maxParallelChecks bounds concurrent outbound health checks.
const maxParallelChecks = 8

// ServiceInput carries the editable fields of a service.
type ServiceInput struct {
	Name, URL, CheckURL, Icon, Description string
	SkipTLSVerify                          bool
}

// ServiceView is a service with its latest status.
type ServiceView struct {
	ID, Name, URL, CheckURL, Icon, Description string
	SkipTLSVerify                              bool
	Position                                   int
	State                                      string
	HTTPStatus                                 int
	LatencyMs                                  int64
	CheckedAt                                  time.Time
	Error                                      string
}

// Services is the input port of the services module.
type Services interface {
	List(ctx context.Context) ([]ServiceView, error)
	Create(ctx context.Context, in ServiceInput) (ServiceView, error)
	Update(ctx context.Context, id string, in ServiceInput) (ServiceView, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) error
	CheckAll(ctx context.Context) error
}

// Repository persists services (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Service, error)
	Get(ctx context.Context, id string) (domain.Service, error)
	Create(ctx context.Context, s domain.Service) error
	Update(ctx context.Context, s domain.Service) error
	Delete(ctx context.Context, id string) error
}

// StatusStore caches the latest health check results (output port).
type StatusStore interface {
	Get(serviceID string) (domain.Status, bool)
	Put(s domain.Status)
	Delete(serviceID string)
}

// HealthChecker performs one HTTP health check (output port).
type HealthChecker interface {
	Check(ctx context.Context, url string, skipTLSVerify bool) (httpStatus int, latency time.Duration, err error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo    Repository
	Status  StatusStore
	Checker HealthChecker
	Clock   kernel.Clock
	IDs     kernel.IDGenerator
}

type services struct{ d Deps }

// NewServices returns the services use cases.
func NewServices(d Deps) Services { return &services{d: d} }

// List implements Services.
func (uc *services) List(ctx context.Context) ([]ServiceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	out := make([]ServiceView, 0, len(list))
	for _, s := range list {
		out = append(out, uc.view(s))
	}
	return out, nil
}

// Create implements Services; the new tile is checked immediately.
func (uc *services) Create(ctx context.Context, in ServiceInput) (ServiceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return ServiceView{}, fmt.Errorf("list services: %w", err)
	}
	s, err := domain.NewService(uc.d.IDs.NewID(), in.Name, in.URL, in.CheckURL, in.Icon, in.Description, in.SkipTLSVerify, len(list))
	if err != nil {
		return ServiceView{}, err
	}
	if err := uc.d.Repo.Create(ctx, s); err != nil {
		return ServiceView{}, fmt.Errorf("create service: %w", err)
	}
	uc.check(ctx, s)
	return uc.view(s), nil
}

// Update implements Services.
func (uc *services) Update(ctx context.Context, id string, in ServiceInput) (ServiceView, error) {
	old, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return ServiceView{}, err
	}
	s, err := domain.NewService(id, in.Name, in.URL, in.CheckURL, in.Icon, in.Description, in.SkipTLSVerify, old.Position)
	if err != nil {
		return ServiceView{}, err
	}
	if err := uc.d.Repo.Update(ctx, s); err != nil {
		return ServiceView{}, fmt.Errorf("update service: %w", err)
	}
	uc.check(ctx, s)
	return uc.view(s), nil
}

// Delete implements Services.
func (uc *services) Delete(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	uc.d.Status.Delete(id)
	return nil
}

// Reorder implements Services; ids must list every service exactly once.
func (uc *services) Reorder(ctx context.Context, ids []string) error {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	byID := make(map[string]domain.Service, len(list))
	for _, s := range list {
		byID[s.ID] = s
	}
	if len(ids) != len(list) {
		return kernel.NewValidationError("ids", "The new order must contain every service exactly once.")
	}
	for i, id := range ids {
		s, ok := byID[id]
		if !ok {
			return kernel.NewValidationError("ids", "The new order must contain every service exactly once.")
		}
		delete(byID, id)
		s.Position = i
		if err := uc.d.Repo.Update(ctx, s); err != nil {
			return fmt.Errorf("update position: %w", err)
		}
	}
	return nil
}

// CheckAll implements Services; it is run by the background job.
func (uc *services) CheckAll(ctx context.Context) error {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	sem := make(chan struct{}, maxParallelChecks)
	var wg sync.WaitGroup
	for _, s := range list {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			uc.check(ctx, s)
		}()
	}
	wg.Wait()
	return nil
}

func (uc *services) check(ctx context.Context, s domain.Service) {
	code, latency, err := uc.d.Checker.Check(ctx, s.CheckURL, s.SkipTLSVerify)
	uc.d.Status.Put(domain.StatusFromResponse(s.ID, code, latency, err, uc.d.Clock.Now()))
}

func (uc *services) view(s domain.Service) ServiceView {
	v := ServiceView{
		ID: s.ID, Name: s.Name, URL: s.URL, CheckURL: s.CheckURL, Icon: s.Icon, Description: s.Description,
		SkipTLSVerify: s.SkipTLSVerify, Position: s.Position, State: string(domain.StateUnknown),
	}
	if st, ok := uc.d.Status.Get(s.ID); ok {
		v.State, v.HTTPStatus, v.LatencyMs, v.CheckedAt, v.Error = string(st.State), st.HTTPStatus, st.Latency.Milliseconds(), st.CheckedAt, st.Error
	}
	return v
}
