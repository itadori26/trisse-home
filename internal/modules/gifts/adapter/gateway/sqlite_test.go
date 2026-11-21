package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/gifts/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestIdeasStayPrivate proves that a second account, admins included, can
// neither read, change nor delete an idea (#684).
func TestIdeasStayPrivate(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "gifts", Migrations()))
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	idea, _ := domain.New("g1", "anna", domain.Input{Person: "Ben", Title: "Headphones", Note: "Over-ear"}, time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC))
	if err := repo.Save(anna, idea); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(anna, "g1")
	if err != nil || got.Note != "Over-ear" || got.Owner != "anna" {
		t.Fatalf("%+v %v", got, err)
	}
	for name, ctx := range map[string]context.Context{"ben": ben, "admin": admin} {
		if list, _ := repo.List(ctx); len(list) != 0 {
			t.Errorf("%s lists anna's idea", name)
		}
		if _, err := repo.Get(ctx, "g1"); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s reads anna's idea: %v", name, err)
		}
		changed := idea
		changed.Title = "Socks"
		if err := repo.Save(ctx, changed); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s changes anna's idea: %v", name, err)
		}
		if err := repo.Delete(ctx, "g1"); err != nil {
			t.Errorf("%s delete: %v", name, err)
		}
	}
	if got, err := repo.Get(anna, "g1"); err != nil || got.Title != "Headphones" {
		t.Fatalf("idea changed: %+v %v", got, err)
	}
	if err := repo.Delete(anna, "g1"); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(anna); len(list) != 0 {
		t.Fatal("deleted idea still listed")
	}
}
