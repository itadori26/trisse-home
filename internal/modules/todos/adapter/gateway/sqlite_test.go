package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "todos", Migrations()))
	now := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	a, _ := domain.NewTodo("a", "with due", now.AddDate(0, 0, 3), 0, now)
	b, _ := domain.NewTodo("b", "done", time.Time{}, 1, now)
	b.SetDone(true, now)
	for _, td := range []domain.Todo{a, b} {
		if err := repo.Save(ctx, td); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.Get(ctx, "a")
	if err != nil || !got.Due().Equal(a.Due()) || got.Done() {
		t.Fatalf("round trip a: %+v %v", got, err)
	}
	if got, _ := repo.Get(ctx, "b"); !got.Done() || !got.Due().IsZero() {
		t.Fatalf("round trip b: %+v", got)
	}
	if err := repo.Delete(ctx, "a", "b"); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ctx); len(list) != 0 {
		t.Fatalf("not deleted: %v", list)
	}
}

func TestRepositoryKeepsScheduleAndReminder(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "todos", Migrations()))
	now := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	td, _ := domain.NewTodo("a", "water plants", now, 0, now)
	rep, _ := domain.NewRecurrence(domain.RepeatWeekly, 2, []int{1, 4})
	rem, _ := domain.NewReminder("07:45", 1)
	if err := td.Schedule(rep, rem); err != nil {
		t.Fatal(err)
	}
	td.MarkReminded()
	if err := repo.Save(ctx, td); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if r := got.Repeat(); r.Kind() != domain.RepeatWeekly || r.Every() != 2 || len(r.Days()) != 2 || r.Days()[1] != 4 {
		t.Fatalf("repeat %+v", r)
	}
	if got.Reminder().At() != "07:45" || got.Reminder().DaysBefore() != 1 || !got.RemindedFor().Equal(td.Due()) {
		t.Fatalf("reminder %+v %v", got.Reminder(), got.RemindedFor())
	}
	// Clearing works through the upsert as well.
	_ = got.Schedule(domain.Recurrence{}, domain.Reminder{})
	got.Bump()
	if err := repo.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(ctx, "a"); got.Repeat().Repeats() || got.Reminder().Set() {
		t.Fatalf("not cleared: %+v", got)
	}
}

// TestRepositoryVersions: a change is stored only on top of the version it
// was made from (#80).
func TestRepositoryVersions(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "todos", Migrations()))
	now := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	td, _ := domain.NewTodo("v", "Milch", time.Time{}, 0, now)
	if err := repo.Save(ctx, td); err != nil {
		t.Fatal(err)
	}
	a, _ := repo.Get(ctx, "v")
	b, _ := repo.Get(ctx, "v")
	if a.Version() != 1 {
		t.Fatalf("new: %d", a.Version())
	}
	_ = a.Edit("Milch kaufen", time.Time{})
	a.Bump()
	if err := repo.Save(ctx, a); err != nil {
		t.Fatal(err)
	}
	// b was loaded before a was stored: its change is refused.
	_ = b.Edit("Hafermilch", time.Time{})
	b.Bump()
	if err := repo.Save(ctx, b); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("lost update: %v", err)
	}
	if got, _ := repo.Get(ctx, "v"); got.Title() != "Milch kaufen" || got.Version() != 2 {
		t.Fatalf("%q v%d", got.Title(), got.Version())
	}
}
