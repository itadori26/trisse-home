package gateway

import (
	"context"
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
