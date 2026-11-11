package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/docker/domain"
)

// ContainerController restarts containers (output port, write access).
type ContainerController interface {
	Restart(ctx context.Context, name string) error
}

// LogSource reads the last lines of a container's log (output port).
type LogSource interface {
	Logs(ctx context.Context, name string, lines int) ([]string, error)
}

// Capabilities tell the UI which actions are available.
type Capabilities struct {
	Enabled    bool
	Containers []string
}

// Actions is the input port for the opt-in container actions.
type Actions interface {
	Capabilities() Capabilities
	Restart(ctx context.Context, name string) error
	Logs(ctx context.Context, name string) ([]string, error)
}

type actions struct {
	policy     domain.ActionPolicy
	controller ContainerController
	logs       LogSource
	refresh    func(context.Context) error
}

// NewActions returns the container actions; refresh updates the cached list
// after a restart.
func NewActions(policy domain.ActionPolicy, controller ContainerController, logs LogSource, refresh func(context.Context) error) Actions {
	return &actions{policy: policy, controller: controller, logs: logs, refresh: refresh}
}

// Capabilities implements Actions.
func (uc *actions) Capabilities() Capabilities {
	return Capabilities{Enabled: uc.policy.Enabled, Containers: uc.policy.AllowedNames()}
}

// Restart implements Actions.
func (uc *actions) Restart(ctx context.Context, name string) error {
	if err := uc.policy.Check(name); err != nil {
		return err
	}
	if err := uc.controller.Restart(ctx, name); err != nil {
		return fmt.Errorf("restart %s: %w", name, err)
	}
	_ = uc.refresh(ctx) // a failed refresh is stored in the cached list
	return nil
}

// Logs implements Actions.
func (uc *actions) Logs(ctx context.Context, name string) ([]string, error) {
	if err := uc.policy.Check(name); err != nil {
		return nil, err
	}
	lines, err := uc.logs.Logs(ctx, name, domain.MaxLogLines)
	if err != nil {
		return nil, fmt.Errorf("logs of %s: %w", name, err)
	}
	return lines, nil
}
