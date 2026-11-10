package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
)

type memRepo struct{ m map[string]domain.Todo }

func (r *memRepo) List(context.Context) ([]domain.Todo, error) {
	var out []domain.Todo
	for _, t := range r.m {
		out = append(out, t)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Todo, error) {
	t, ok := r.m[id]
	if !ok {
		return t, domain.ErrTodoNotFound
	}
	return t, nil
}
func (r *memRepo) Save(_ context.Context, t domain.Todo) error { r.m[t.ID()] = t; return nil }
func (r *memRepo) Delete(_ context.Context, ids ...string) error {
	for _, id := range ids {
		delete(r.m, id)
	}
	return nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint(i.n) }

func TestTodos(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)}
	d := Deps{Repo: &memRepo{m: map[string]domain.Todo{}}, Clock: c, IDs: &ids{}}
	create := NewCreateTodo(d)
	a, _ := create.Execute(ctx, TodoInput{Title: "first"})
	b, _ := create.Execute(ctx, TodoInput{Title: "second", Due: c.t.AddDate(0, 0, -1)})
	if b.Position != a.Position+1 || !b.Overdue {
		t.Fatalf("unexpected second todo %+v", b)
	}
	if _, err := create.Execute(ctx, TodoInput{Title: ""}); err == nil {
		t.Fatal("expected validation error")
	}
	if _, err := NewUpdateTodo(d).Execute(ctx, a.ID, TodoInput{Title: "first"}, true); err != nil {
		t.Fatal(err)
	}
	open, _ := NewListTodos(d).Execute(ctx, false)
	all, _ := NewListTodos(d).Execute(ctx, true)
	if len(open) != 1 || len(all) != 2 || all[1].ID != a.ID {
		t.Fatalf("listing wrong: open=%v all=%v", open, all)
	}
	if n, err := NewClearCompleted(d).Execute(ctx); err != nil || n != 1 {
		t.Fatalf("clear completed: %d %v", n, err)
	}
	if err := NewDeleteTodo(d).Execute(ctx, a.ID); !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
