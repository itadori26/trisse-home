package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"waylight/internal/modules/gifts/domain"
	"waylight/internal/shared/kernel"
)

// Gifts is the input port of the module; every method is one use case.
// Every method works on the principal's own ideas only.
type Gifts interface {
	// List returns the principal's ideas by person, open ones first.
	List(ctx context.Context) ([]domain.Idea, error)
	Create(ctx context.Context, in domain.Input) (domain.Idea, error)
	Update(ctx context.Context, id string, in domain.Input) (domain.Idea, error)
	Delete(ctx context.Context, id string) error
}

// Repository persists ideas (output port, implemented in adapter/gateway).
// It reads and changes only the ideas of the principal in ctx.
type Repository interface {
	List(ctx context.Context) ([]domain.Idea, error)
	Get(ctx context.Context, id string) (domain.Idea, error)
	Save(ctx context.Context, i domain.Idea) error
	Delete(ctx context.Context, id string) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
	Clock kernel.Clock
	IDs   kernel.IDGenerator
}

type uc struct{ d Deps }

// NewGifts returns the gift idea use cases.
func NewGifts(d Deps) Gifts { return &uc{d: d} }

// List implements Gifts.
func (u *uc) List(ctx context.Context) ([]domain.Idea, error) {
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list gift ideas: %w", err)
	}
	slices.SortStableFunc(list, func(a, b domain.Idea) int {
		if c := strings.Compare(strings.ToLower(a.Person), strings.ToLower(b.Person)); c != 0 {
			return c
		}
		if a.Bought != b.Bought {
			if a.Bought {
				return 1
			}
			return -1
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	return list, nil
}

// Create implements Gifts.
func (u *uc) Create(ctx context.Context, in domain.Input) (domain.Idea, error) {
	i, err := domain.New(u.d.IDs.NewID(), kernel.PrincipalFrom(ctx).AccountID, in, u.d.Clock.Now())
	if err != nil {
		return domain.Idea{}, err
	}
	if err := u.d.Repo.Save(ctx, i); err != nil {
		return domain.Idea{}, fmt.Errorf("save gift idea: %w", err)
	}
	return i, nil
}

// Update implements Gifts.
func (u *uc) Update(ctx context.Context, id string, in domain.Input) (domain.Idea, error) {
	i, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return domain.Idea{}, err
	}
	if i, err = i.Update(in); err != nil {
		return domain.Idea{}, err
	}
	if err := u.d.Repo.Save(ctx, i); err != nil {
		return domain.Idea{}, err
	}
	return i, nil
}

// Delete implements Gifts.
func (u *uc) Delete(ctx context.Context, id string) error { return u.d.Repo.Delete(ctx, id) }

// OpenCount is the number of ideas not bought yet, for the compact state.
func OpenCount(list []domain.Idea) int {
	n := 0
	for _, i := range list {
		if !i.Bought {
			n++
		}
	}
	return n
}
