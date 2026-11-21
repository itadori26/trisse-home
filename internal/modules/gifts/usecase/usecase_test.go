package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/gifts/domain"
	"waylight/internal/shared/kernel"
)

// memRepo keeps ideas in memory, each visible to its owner only, like the
// SQLite gateway.
type memRepo struct{ ideas map[string]domain.Idea }

func owner(ctx context.Context) string { return kernel.PrincipalFrom(ctx).AccountID }

func (m *memRepo) List(ctx context.Context) ([]domain.Idea, error) {
	var out []domain.Idea
	for _, i := range m.ideas {
		if i.Owner == owner(ctx) {
			out = append(out, i)
		}
	}
	return out, nil
}

func (m *memRepo) Get(ctx context.Context, id string) (domain.Idea, error) {
	i, ok := m.ideas[id]
	if !ok || i.Owner != owner(ctx) {
		return domain.Idea{}, domain.ErrNotFound
	}
	return i, nil
}

func (m *memRepo) Save(_ context.Context, i domain.Idea) error { m.ideas[i.ID] = i; return nil }

func (m *memRepo) Delete(ctx context.Context, id string) error {
	if m.ideas[id].Owner == owner(ctx) {
		delete(m.ideas, id)
	}
	return nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "g" + strconv.Itoa(i.n) }

var (
	anna = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "admin"})
)

func TestIdeasArePrivate(t *testing.T) {
	g := NewGifts(Deps{Repo: &memRepo{ideas: map[string]domain.Idea{}}, Clock: clk{}, IDs: &ids{}})
	i, err := g.Create(anna, domain.Input{Person: "Ben", Title: "Headphones"})
	if err != nil || i.Owner != "anna" {
		t.Fatalf("%+v %v", i, err)
	}
	if list, _ := g.List(ben); len(list) != 0 {
		t.Fatalf("ben sees anna's ideas about him: %v", list)
	}
	if _, err := g.Update(ben, i.ID, domain.Input{Person: "Ben", Title: "Socks"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben changed anna's idea: %v", err)
	}
}

func TestListGroupsByPersonOpenFirst(t *testing.T) {
	g := NewGifts(Deps{Repo: &memRepo{ideas: map[string]domain.Idea{}}, Clock: clk{}, IDs: &ids{}})
	for _, in := range []domain.Input{
		{Person: "paul", Title: "Bought", Bought: true},
		{Person: "Anna", Title: "Book"},
		{Person: "Paul", Title: "Ball"},
	} {
		if _, err := g.Create(anna, in); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := g.List(anna)
	if len(list) != 3 || list[0].Title != "Book" || list[1].Title != "Ball" || list[2].Title != "Bought" || OpenCount(list) != 2 {
		t.Fatalf("%+v", list)
	}
}

func TestSeedDemoOnce(t *testing.T) {
	repo := &memRepo{ideas: map[string]domain.Idea{}}
	d := Deps{Repo: repo, Clock: clk{}, IDs: &ids{}}
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if err := SeedDemo(ctx, d); err != nil || len(repo.ideas) == 0 {
		t.Fatalf("%v %d", err, len(repo.ideas))
	}
	n := len(repo.ideas)
	if err := SeedDemo(ctx, d); err != nil || len(repo.ideas) != n {
		t.Fatalf("second seed: %d → %d %v", n, len(repo.ideas), err)
	}
}
