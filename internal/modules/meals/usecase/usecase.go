package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/meals/domain"
	"waylight/internal/shared/kernel"
)

// Meals is the input port of the module; every method is one use case.
type Meals interface {
	// Range lists the visible dishes of from (YYYY-MM-DD, empty = today)
	// and the following days.
	Range(ctx context.Context, from string, days int) ([]EntryView, error)
	// Create plans a dish. An empty visibility shares it with everyone.
	Create(ctx context.Context, in domain.Input, visibility string) (EntryView, error)
	// Update changes a dish of the principal; an empty visibility keeps it.
	Update(ctx context.Context, id string, in domain.Input, visibility string) (EntryView, error)
	Delete(ctx context.Context, id string) error
}

// EntryView is a dish as the principal sees it.
type EntryView struct {
	domain.Entry
	// Mine says whether the principal created the dish and may change it.
	Mine bool
}

// Repository persists dishes (output port, implemented in adapter/gateway).
// Reads return only what the principal in ctx may see; Save and Delete
// change only the principal's own entries (ADR 0018).
type Repository interface {
	Between(ctx context.Context, from, to time.Time) ([]domain.Entry, error)
	Get(ctx context.Context, id string) (domain.Entry, error)
	Save(ctx context.Context, e domain.Entry) error
	Delete(ctx context.Context, id string) error
}

// Deps bundles the output ports. Location is the household's time zone
// that decides what "today" is (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewMeals returns the meal plan use cases.
func NewMeals(d Deps) Meals {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) view(ctx context.Context, e domain.Entry) EntryView {
	return EntryView{Entry: e, Mine: e.Access.Mine(kernel.PrincipalFrom(ctx))}
}

// Range implements Meals.
func (u *uc) Range(ctx context.Context, from string, days int) ([]EntryView, error) {
	r, err := domain.NewRange(from, days, u.d.Clock.Now().In(u.d.Location))
	if err != nil {
		return nil, err
	}
	list, err := u.d.Repo.Between(ctx, r.From, r.End())
	if err != nil {
		return nil, fmt.Errorf("list dishes: %w", err)
	}
	slices.SortStableFunc(list, func(a, b domain.Entry) int {
		switch {
		case domain.Less(a, b):
			return -1
		case domain.Less(b, a):
			return 1
		}
		return 0
	})
	out := make([]EntryView, 0, len(list))
	for _, e := range list {
		out = append(out, u.view(ctx, e))
	}
	return out, nil
}

// shareFor is the visibility of a new dish: everyone unless the creator
// chose otherwise, because a meal plan is meant for the household.
func shareFor(p kernel.Principal, raw string) (kernel.Visibility, error) {
	if raw == "" {
		return kernel.Everyone, nil
	}
	return kernel.CheckShare(p, raw)
}

// Create implements Meals.
func (u *uc) Create(ctx context.Context, in domain.Input, visibility string) (EntryView, error) {
	p := kernel.PrincipalFrom(ctx)
	v, err := shareFor(p, visibility)
	if err != nil {
		return EntryView{}, err
	}
	e, err := domain.New(u.d.IDs.NewID(), in, kernel.Access{Owner: p.AccountID, Visibility: v}, u.d.Clock.Now())
	if err != nil {
		return EntryView{}, err
	}
	if err := u.d.Repo.Save(ctx, e); err != nil {
		return EntryView{}, fmt.Errorf("save dish: %w", err)
	}
	return u.view(ctx, e), nil
}

// Update implements Meals.
func (u *uc) Update(ctx context.Context, id string, in domain.Input, visibility string) (EntryView, error) {
	p := kernel.PrincipalFrom(ctx)
	e, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return EntryView{}, err
	}
	if !e.Access.Mine(p) {
		return EntryView{}, kernel.ErrNotOwner
	}
	if visibility != "" {
		if e.Access.Visibility, err = kernel.CheckShare(p, visibility); err != nil {
			return EntryView{}, err
		}
	}
	if e, err = e.Update(in); err != nil {
		return EntryView{}, err
	}
	if err := u.d.Repo.Save(ctx, e); err != nil {
		return EntryView{}, err
	}
	return u.view(ctx, e), nil
}

// Delete implements Meals.
func (u *uc) Delete(ctx context.Context, id string) error {
	return u.d.Repo.Delete(ctx, id)
}

// Today returns the visible dishes of today, for the compact state.
func Today(ctx context.Context, m Meals) ([]EntryView, error) { return m.Range(ctx, "", 1) }
