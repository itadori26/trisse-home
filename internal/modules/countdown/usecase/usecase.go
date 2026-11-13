package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"trisse-home/internal/modules/countdown/domain"
	"trisse-home/internal/shared/kernel"
)

// Countdowns is the input port of the module; every method is one use
// case (#54).
type Countdowns interface {
	// List returns the countdowns sorted by their end; past ones stay
	// visible for graceDays days.
	List(ctx context.Context, graceDays int) ([]domain.Countdown, error)
	Create(ctx context.Context, in domain.Input) (domain.Countdown, error)
	Update(ctx context.Context, id string, in domain.Input) (domain.Countdown, error)
	Delete(ctx context.Context, id string) error
}

// Repository persists countdowns (output port, implemented in
// adapter/gateway).
type Repository interface {
	List(ctx context.Context) ([]domain.Countdown, error)
	Get(ctx context.Context, id string) (domain.Countdown, error)
	Save(ctx context.Context, c domain.Countdown) error
	Delete(ctx context.Context, id string) error
}

// Deps bundles the output ports. Location is the user's time zone
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewCountdowns returns the countdown use cases.
func NewCountdowns(d Deps) Countdowns {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) now() time.Time { return u.d.Clock.Now().In(u.d.Location) }

// List implements Countdowns.
func (u *uc) List(ctx context.Context, graceDays int) ([]domain.Countdown, error) {
	all, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list countdowns: %w", err)
	}
	now, grace := u.now(), domain.ClampGrace(graceDays)
	out := make([]domain.Countdown, 0, len(all))
	for _, c := range all {
		if !c.Hidden(now, grace) {
			out = append(out, c)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].TargetIn(u.d.Location), out[j].TargetIn(u.d.Location)
		if !a.Equal(b) {
			return a.Before(b)
		}
		return out[i].Title < out[j].Title
	})
	return out, nil
}

// Create implements Countdowns.
func (u *uc) Create(ctx context.Context, in domain.Input) (domain.Countdown, error) {
	c, err := domain.New(u.d.IDs.NewID(), in, u.d.Clock.Now())
	if err != nil {
		return domain.Countdown{}, err
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return domain.Countdown{}, fmt.Errorf("save countdown: %w", err)
	}
	return c, nil
}

// Update implements Countdowns.
func (u *uc) Update(ctx context.Context, id string, in domain.Input) (domain.Countdown, error) {
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return domain.Countdown{}, err
	}
	if c, err = c.Update(in); err != nil {
		return domain.Countdown{}, err
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return domain.Countdown{}, fmt.Errorf("save countdown: %w", err)
	}
	return c, nil
}

// Delete implements Countdowns.
func (u *uc) Delete(ctx context.Context, id string) error {
	if _, err := u.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	return u.d.Repo.Delete(ctx, id)
}
