package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/chores/domain"
	"waylight/internal/shared/kernel"
)

// memRepo keeps chores in memory and applies the visibility rule like the
// SQLite gateway.
type memRepo struct{ chores map[string]domain.Chore }

func (m *memRepo) visible(ctx context.Context, c domain.Chore) bool {
	return c.Access.Visibility.VisibleTo(kernel.PrincipalFrom(ctx), c.Access.Owner)
}

func (m *memRepo) List(ctx context.Context) ([]domain.Chore, error) {
	var out []domain.Chore
	for _, c := range m.chores {
		if m.visible(ctx, c) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *memRepo) Get(ctx context.Context, id string) (domain.Chore, error) {
	c, ok := m.chores[id]
	if !ok || !m.visible(ctx, c) {
		return domain.Chore{}, domain.ErrNotFound
	}
	return c, nil
}

func (m *memRepo) Save(_ context.Context, c domain.Chore) error { m.chores[c.ID] = c; return nil }
func (m *memRepo) Delete(_ context.Context, id string) error    { delete(m.chores, id); return nil }

func (m *memRepo) SaveDone(_ context.Context, c domain.Chore, prevDue time.Time) (bool, error) {
	if !m.chores[c.ID].Due.Equal(prevDue) {
		return false, nil
	}
	m.chores[c.ID] = c
	return true, nil
}

type clk struct{ now time.Time }

func (c clk) Now() time.Time { return c.now }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "c" + strconv.Itoa(i.n) }

var (
	now  = time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)
	anna = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
)

func setup() (Chores, *memRepo) {
	repo := &memRepo{chores: map[string]domain.Chore{}}
	return NewChores(Deps{Repo: repo, Clock: clk{now}, IDs: &ids{}, Location: time.UTC}), repo
}

func TestListSortsByDue(t *testing.T) {
	c, _ := setup()
	for _, in := range []domain.Input{
		{Title: "Later", EveryDays: 7, Due: "2026-11-25"},
		{Title: "overdue", EveryDays: 7, Due: "2026-11-20"},
		{Title: "Today", EveryDays: 1, Due: "2026-11-23"},
	} {
		if _, err := c.Create(anna, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	list, err := c.List(ben)
	if err != nil || len(list) != 3 {
		t.Fatalf("%v %v", list, err)
	}
	if list[0].Title != "overdue" || !list[0].Overdue || list[1].Title != "Today" || !list[1].DueNow || list[1].Overdue || list[2].DueNow || list[0].Mine {
		t.Fatalf("%+v", list)
	}
}

func TestEveryoneWhoSeesMayTickButOnlyOwnerEdits(t *testing.T) {
	c, _ := setup()
	v, _ := c.Create(anna, domain.Input{Title: "Dishes", EveryDays: 1, Due: "2026-11-23", Assignees: []string{"Anna", "Ben"}}, "")
	done, err := c.Done(ben, v.ID)
	if err != nil || done.DueDate() != "2026-11-24" || done.Assignee() != "Ben" || done.LastDoneBy != "Anna" {
		t.Fatalf("%+v %v", done, err)
	}
	if _, err := c.Update(ben, v.ID, domain.Input{Title: "Mine", EveryDays: 1, Due: "2026-11-23"}, ""); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's chore: %v", err)
	}
	priv, _ := c.Create(anna, domain.Input{Title: "Secret", EveryDays: 1, Due: "2026-11-23"}, "private")
	if _, err := c.Done(ben, priv.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben ticked a private chore: %v", err)
	}
}

func TestSeedDemoOnce(t *testing.T) {
	repo := &memRepo{chores: map[string]domain.Chore{}}
	d := Deps{Repo: repo, Clock: clk{now}, IDs: &ids{}, Location: time.UTC}
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if err := SeedDemo(ctx, d); err != nil || len(repo.chores) == 0 {
		t.Fatalf("%v %d", err, len(repo.chores))
	}
	n := len(repo.chores)
	if err := SeedDemo(ctx, d); err != nil || len(repo.chores) != n {
		t.Fatalf("second seed: %d → %d %v", n, len(repo.chores), err)
	}
}
