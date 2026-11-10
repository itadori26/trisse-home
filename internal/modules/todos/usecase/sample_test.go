package usecase

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
)

func TestSeedSampleData(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{m: map[string]domain.Todo{}}
	d := Deps{Repo: repo, Clock: &clk{t: time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)}, IDs: &ids{}}
	for range 2 {
		if err := NewSeedSampleData(d).Execute(ctx); err != nil {
			t.Fatal(err)
		}
	}
	open, _ := NewListTodos(d).Execute(ctx, false)
	all, _ := NewListTodos(d).Execute(ctx, true)
	if len(open) != 4 || len(all) != 5 {
		t.Fatalf("expected 4 open and 5 total to-dos, got %d/%d", len(open), len(all))
	}
}
