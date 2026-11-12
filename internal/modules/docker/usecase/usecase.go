package usecase

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/shared/kernel"
)

// ContainerSource lists containers (output port).
type ContainerSource interface {
	Containers(ctx context.Context) ([]domain.Container, error)
}

// ContainerView is a container at the use case boundary.
type ContainerView struct {
	ID, Name, Image, State, Status, Health, Project string
	Created                                         time.Time
	Labels                                          map[string]string
}

// ListResponse is the cached container list.
type ListResponse struct {
	Containers                         []ContainerView
	Total, Running, Stopped, Unhealthy int
	UpdatedAt                          time.Time
	// Error is set when the last refresh failed; Containers then holds the
	// last known state.
	Error string
}

// Containers is the input port of the docker module.
type Containers interface {
	Refresh(ctx context.Context) error
	List(ctx context.Context) (ListResponse, error)
}

type containers struct {
	source ContainerSource
	sink   kernel.ObservationSink
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *ListResponse
}

// NewContainers returns the container use cases; every refresh reports the
// state of each container to sink.
func NewContainers(source ContainerSource, sink kernel.ObservationSink, clock kernel.Clock) Containers {
	return &containers{source: source, sink: sink, clock: clock}
}

// Refresh implements Containers; it is run by the background job.
func (uc *containers) Refresh(ctx context.Context) error {
	list, err := uc.source.Containers(ctx)
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if err != nil {
		res := ListResponse{Error: "The Docker API is not reachable.", UpdatedAt: uc.clock.Now()}
		if uc.cached != nil {
			res = *uc.cached
			res.Error = "The Docker API is not reachable; showing the last known state."
		}
		uc.cached = &res
		return fmt.Errorf("list containers: %w", err)
	}
	for _, c := range list {
		uc.sink.Observe(ctx, kernel.Observation{Kind: "container", Key: c.Name, Label: c.Name,
			Healthy: c.State == "running" && c.Health() != domain.HealthUnhealthy, Detail: c.Status})
	}
	domain.SortForDisplay(list)
	s := domain.Summarise(list)
	res := ListResponse{Total: s.Total, Running: s.Running, Stopped: s.Stopped, Unhealthy: s.Unhealthy, UpdatedAt: uc.clock.Now()}
	for _, c := range list {
		res.Containers = append(res.Containers, ContainerView{
			ID: c.ID, Name: c.Name, Image: c.Image, State: c.State, Status: c.Status,
			Health: string(c.Health()), Project: c.Project, Created: c.Created, Labels: c.Labels,
		})
	}
	uc.cached = &res
	return nil
}

// List implements Containers (cache-aside).
func (uc *containers) List(ctx context.Context) (ListResponse, error) {
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached == nil {
		_ = uc.Refresh(ctx) // the error is stored in the cached response
		uc.mu.RLock()
		cached = uc.cached
		uc.mu.RUnlock()
	}
	res := *cached
	res.Containers = slices.Clone(res.Containers)
	return res, nil
}
