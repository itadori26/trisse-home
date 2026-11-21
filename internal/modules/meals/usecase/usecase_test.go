package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/meals/domain"
	"waylight/internal/shared/kernel"
)

// memRepo keeps dishes in memory and applies the visibility rule like the
// SQLite gateway, so the use cases can be tested without a database.
type memRepo struct{ entries map[string]domain.Entry }

func (m *memRepo) Between(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	p := kernel.PrincipalFrom(ctx)
	var out []domain.Entry
	for _, e := range m.entries {
		if !e.Day.Before(from) && e.Day.Before(to) && e.Access.Visibility.VisibleTo(p, e.Access.Owner) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *memRepo) Get(ctx context.Context, id string) (domain.Entry, error) {
	e, ok := m.entries[id]
	if !ok || !e.Access.Visibility.VisibleTo(kernel.PrincipalFrom(ctx), e.Access.Owner) {
		return domain.Entry{}, domain.ErrNotFound
	}
	return e, nil
}

func (m *memRepo) Save(_ context.Context, e domain.Entry) error { m.entries[e.ID] = e; return nil }
func (m *memRepo) Delete(_ context.Context, id string) error    { delete(m.entries, id); return nil }

type clk struct{ now time.Time }

func (c clk) Now() time.Time { return c.now }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "e" + strconv.Itoa(i.n) }

func setup() (Meals, *memRepo) {
	repo := &memRepo{entries: map[string]domain.Entry{}}
	now := time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)
	return NewMeals(Deps{Repo: repo, Clock: clk{now}, IDs: &ids{}, Location: time.UTC}), repo
}

var (
	anna = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member", Groups: []string{"kids"}})
	ben  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
)

func TestCreateSharesWithEveryoneByDefault(t *testing.T) {
	m, _ := setup()
	v, err := m.Create(anna, domain.Input{Date: "2026-11-23", Dish: "Soup"}, "")
	if err != nil || v.Access.Visibility != kernel.Everyone || !v.Mine {
		t.Fatalf("%+v %v", v, err)
	}
	if _, err := m.Create(anna, domain.Input{Date: "2026-11-23", Dish: "Soup"}, "group:work"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("shared with a foreign group: %v", err)
	}
	list, err := m.Range(ben, "", 1)
	if err != nil || len(list) != 1 || list[0].Mine {
		t.Fatalf("ben sees %+v %v", list, err)
	}
}

func TestRangeSortsByDayAndMeal(t *testing.T) {
	m, _ := setup()
	for _, in := range []domain.Input{
		{Date: "2026-11-24", Meal: "lunch", Dish: "Tomorrow"},
		{Date: "2026-11-23", Meal: "dinner", Dish: "Dinner"},
		{Date: "2026-11-23", Meal: "breakfast", Dish: "Breakfast"},
		{Date: "2026-11-30", Dish: "Next week"},
	} {
		if _, err := m.Create(anna, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	list, err := m.Range(anna, "2026-11-23", 7)
	if err != nil || len(list) != 3 || list[0].Dish != "Breakfast" || list[1].Dish != "Dinner" || list[2].Dish != "Tomorrow" {
		t.Fatalf("%+v %v", list, err)
	}
	if _, err := m.Range(anna, "", 0); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("zero days accepted")
	}
}

func TestUpdateOnlyByOwner(t *testing.T) {
	m, _ := setup()
	v, _ := m.Create(anna, domain.Input{Date: "2026-11-23", Dish: "Soup"}, "")
	if _, err := m.Update(ben, v.ID, domain.Input{Date: "2026-11-23", Dish: "Stew"}, ""); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's dish: %v", err)
	}
	got, err := m.Update(anna, v.ID, domain.Input{Date: "2026-11-24", Dish: "Stew"}, "private")
	if err != nil || got.Dish != "Stew" || got.Access.Visibility != kernel.Private || got.Date() != "2026-11-24" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := m.Update(anna, "missing", domain.Input{Date: "2026-11-24", Dish: "Stew"}, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown dish: %v", err)
	}
}

func TestSeedDemoOnce(t *testing.T) {
	repo := &memRepo{entries: map[string]domain.Entry{}}
	d := Deps{Repo: repo, Clock: clk{time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)}, IDs: &ids{}, Location: time.UTC}
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if err := SeedDemo(ctx, d); err != nil {
		t.Fatal(err)
	}
	n := len(repo.entries)
	if n == 0 {
		t.Fatal("no demo dishes")
	}
	if err := SeedDemo(ctx, d); err != nil || len(repo.entries) != n {
		t.Fatalf("second seed added dishes: %d → %d, %v", n, len(repo.entries), err)
	}
}
