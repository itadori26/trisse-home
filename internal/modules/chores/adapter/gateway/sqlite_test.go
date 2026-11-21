package gateway

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"waylight/internal/modules/chores/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

var now = time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)

func chore(t *testing.T, id string, v kernel.Visibility) domain.Chore {
	t.Helper()
	c, err := domain.New(id, domain.Input{Title: "Chore " + id, EveryDays: 7, Due: "2026-11-23", Assignees: []string{"Anna", "Ben"}}, kernel.Access{Visibility: v}, now)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

var (
	anna = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
)

func TestRoundTripAndDone(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "chores", Migrations()))
	c := chore(t, "a", kernel.Everyone)
	if err := repo.Save(anna, c); err != nil {
		t.Fatal(err)
	}
	done := c.Done(now, now)
	if ok, err := repo.SaveDone(ben, done, c.Due); !ok || err != nil {
		t.Fatalf("ben may tick a shared chore: %v %v", ok, err)
	}
	// A second completion based on the old state is refused.
	if ok, err := repo.SaveDone(ben, done.Done(now, now), c.Due); ok || err != nil {
		t.Fatalf("stale completion: %v %v", ok, err)
	}
	got, err := repo.Get(anna, "a")
	if err != nil || got.DueDate() != "2026-11-30" || got.Assignee() != "Ben" || got.LastDoneBy != "Anna" || !got.LastDoneAt.Equal(now) ||
		!slices.Equal(got.Assignees, []string{"Anna", "Ben"}) || got.Access.Owner != "anna" {
		t.Fatalf("%+v %v", got, err)
	}
}

// TestOwnership proves ADR 0018 for chores.
func TestOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "chores", Migrations()))
	shared, private := chore(t, "shared", kernel.Everyone), chore(t, "private", kernel.Private)
	for _, c := range []domain.Chore{shared, private} {
		if err := repo.Save(anna, c); err != nil {
			t.Fatal(err)
		}
	}
	if list, _ := repo.List(ben); len(list) != 1 || list[0].ID != "shared" {
		t.Fatalf("ben sees %v", list)
	}
	if _, err := repo.Get(ben, "private"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben read anna's private chore: %v", err)
	}
	if ok, _ := repo.SaveDone(ben, private.Done(now, now), private.Due); ok {
		t.Fatal("ben ticked anna's private chore")
	}
	shared.Title = "Mine now"
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's chore: %v", err)
	}
	if err := repo.Delete(ben, "shared"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted anna's chore: %v", err)
	}
	if err := repo.Delete(anna, "shared"); err != nil {
		t.Fatal(err)
	}
}
