package gateway

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"waylight/internal/modules/meals/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

var created = time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)

func entry(t *testing.T, id, date string, v kernel.Visibility) domain.Entry {
	t.Helper()
	e, err := domain.New(id, domain.Input{Date: date, Meal: "dinner", Dish: "Dish " + id, Ingredients: []string{"salt", "pepper"}}, kernel.Access{Visibility: v}, created)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func day(s string) time.Time {
	d, _ := time.Parse(time.DateOnly, s)
	return d
}

func TestRepositoryRoundTrip(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "meals", Migrations()))
	ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	for _, e := range []domain.Entry{entry(t, "a", "2026-11-23", kernel.Everyone), entry(t, "b", "2026-11-30", kernel.Everyone)} {
		if err := repo.Save(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.Between(ctx, day("2026-11-23"), day("2026-11-30"))
	if err != nil || len(list) != 1 {
		t.Fatalf("%v %v", list, err)
	}
	got := list[0]
	if got.ID != "a" || got.Meal != domain.Dinner || !slices.Equal(got.Ingredients, []string{"salt", "pepper"}) || got.Access.Owner != "anna" || !got.CreatedAt.Equal(created) {
		t.Fatalf("%+v", got)
	}
	if err := repo.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

// TestOwnership proves ADR 0018 for the meal plan.
func TestOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "meals", Migrations()))
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	shared, private := entry(t, "shared", "2026-11-23", kernel.Everyone), entry(t, "private", "2026-11-23", kernel.Private)
	for _, e := range []domain.Entry{shared, private} {
		if err := repo.Save(anna, e); err != nil {
			t.Fatal(err)
		}
	}
	if list, _ := repo.Between(ben, day("2026-11-23"), day("2026-11-24")); len(list) != 1 || list[0].ID != "shared" {
		t.Fatalf("ben sees %v", list)
	}
	if _, err := repo.Get(ben, "private"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben read anna's private dish: %v", err)
	}
	shared.Dish = "Mine now"
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's dish: %v", err)
	}
	if err := repo.Delete(ben, "shared"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted anna's dish: %v", err)
	}
	if err := repo.Delete(ben, "private"); err != nil {
		t.Fatalf("an invisible dish is no error: %v", err)
	}
	if got, _ := repo.Get(anna, "private"); got.Dish != "Dish private" {
		t.Fatalf("private dish changed: %+v", got)
	}
}
