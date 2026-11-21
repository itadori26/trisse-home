package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"waylight/internal/modules/chores/domain"
	"waylight/internal/shared/kernel"
)

// Chores is the input port of the module; every method is one use case.
type Chores interface {
	// List returns the visible chores, the ones due soonest first.
	List(ctx context.Context) ([]ChoreView, error)
	// Create adds a chore. An empty visibility shares it with everyone.
	Create(ctx context.Context, in domain.Input, visibility string) (ChoreView, error)
	// Update changes a chore of the principal; an empty visibility keeps it.
	Update(ctx context.Context, id string, in domain.Input, visibility string) (ChoreView, error)
	Delete(ctx context.Context, id string) error
	// Done marks a visible chore done today; everyone who sees it may.
	Done(ctx context.Context, id string) (ChoreView, error)
}

// ChoreView is a chore as the principal sees it today.
type ChoreView struct {
	domain.Chore
	// Mine says whether the principal created the chore and may change it.
	Mine    bool
	DueNow  bool
	Overdue bool
}

// Repository persists chores (output port, implemented in adapter/gateway).
// Reads return only what the principal in ctx may see; Save and Delete
// change only the principal's own chores (ADR 0018).
type Repository interface {
	List(ctx context.Context) ([]domain.Chore, error)
	Get(ctx context.Context, id string) (domain.Chore, error)
	Save(ctx context.Context, c domain.Chore) error
	Delete(ctx context.Context, id string) error
	// SaveDone stores a completion of a visible chore, as long as it is
	// still due on prevDue; false when someone else was quicker.
	SaveDone(ctx context.Context, c domain.Chore, prevDue time.Time) (bool, error)
}

// Deps bundles the output ports. Location decides what "today" is
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewChores returns the chore use cases.
func NewChores(d Deps) Chores {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) today() time.Time { return u.d.Clock.Now().In(u.d.Location) }

func (u *uc) view(ctx context.Context, c domain.Chore) ChoreView {
	today := u.today()
	return ChoreView{Chore: c, Mine: c.Access.Mine(kernel.PrincipalFrom(ctx)), DueNow: c.DueOn(today), Overdue: c.OverdueOn(today)}
}

// List implements Chores.
func (u *uc) List(ctx context.Context) ([]ChoreView, error) {
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list chores: %w", err)
	}
	slices.SortStableFunc(list, func(a, b domain.Chore) int {
		if c := a.Due.Compare(b.Due); c != 0 {
			return c
		}
		return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})
	out := make([]ChoreView, 0, len(list))
	for _, c := range list {
		out = append(out, u.view(ctx, c))
	}
	return out, nil
}

// shareFor is the visibility of a new chore: everyone unless the creator
// chose otherwise, because chores are shared by the household.
func shareFor(p kernel.Principal, raw string) (kernel.Visibility, error) {
	if raw == "" {
		return kernel.Everyone, nil
	}
	return kernel.CheckShare(p, raw)
}

// Create implements Chores.
func (u *uc) Create(ctx context.Context, in domain.Input, visibility string) (ChoreView, error) {
	p := kernel.PrincipalFrom(ctx)
	v, err := shareFor(p, visibility)
	if err != nil {
		return ChoreView{}, err
	}
	c, err := domain.New(u.d.IDs.NewID(), in, kernel.Access{Owner: p.AccountID, Visibility: v}, u.d.Clock.Now())
	if err != nil {
		return ChoreView{}, err
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return ChoreView{}, fmt.Errorf("save chore: %w", err)
	}
	return u.view(ctx, c), nil
}

// Update implements Chores.
func (u *uc) Update(ctx context.Context, id string, in domain.Input, visibility string) (ChoreView, error) {
	p := kernel.PrincipalFrom(ctx)
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return ChoreView{}, err
	}
	if !c.Access.Mine(p) {
		return ChoreView{}, kernel.ErrNotOwner
	}
	if visibility != "" {
		if c.Access.Visibility, err = kernel.CheckShare(p, visibility); err != nil {
			return ChoreView{}, err
		}
	}
	if c, err = c.Update(in); err != nil {
		return ChoreView{}, err
	}
	if err := u.d.Repo.Save(ctx, c); err != nil {
		return ChoreView{}, err
	}
	return u.view(ctx, c), nil
}

// Delete implements Chores.
func (u *uc) Delete(ctx context.Context, id string) error { return u.d.Repo.Delete(ctx, id) }

// Done implements Chores. Two taps at once (the widget and the compact
// view) record one completion: the second sees the chore already moved on
// and returns it unchanged.
func (u *uc) Done(ctx context.Context, id string) (ChoreView, error) {
	c, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return ChoreView{}, err
	}
	done := c.Done(u.today(), u.d.Clock.Now())
	ok, err := u.d.Repo.SaveDone(ctx, done, c.Due)
	if err != nil {
		return ChoreView{}, fmt.Errorf("save completion: %w", err)
	}
	if !ok {
		if done, err = u.d.Repo.Get(ctx, id); err != nil {
			return ChoreView{}, err
		}
	}
	return u.view(ctx, done), nil
}
