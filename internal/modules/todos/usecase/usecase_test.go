package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
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

type recordAlerts struct{ got []kernel.Alert }

func (r *recordAlerts) Alert(_ context.Context, a kernel.Alert) { r.got = append(r.got, a) }

func TestRecurringTodosAndReminders(t *testing.T) {
	ctx := context.Background()
	loc := time.FixedZone("CET", 3600)
	// Thursday 12 November 2026, 23:30 local = 22:30 UTC.
	c := &clk{t: time.Date(2026, 11, 12, 22, 30, 0, 0, time.UTC)}
	repo := &memRepo{m: map[string]domain.Todo{}}
	alerts := &recordAlerts{}
	d := Deps{Repo: repo, Clock: c, IDs: &ids{}, Location: loc, Alerts: alerts}
	due := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

	if _, err := NewCreateTodo(d).Execute(ctx, TodoInput{Title: "x", Reminder: &ReminderInput{At: "08:00"}}); err == nil {
		t.Fatal("reminder without due date accepted")
	}
	if _, err := NewCreateTodo(d).Execute(ctx, TodoInput{Title: "x", Due: due, Repeat: &RepeatInput{Kind: "weekly", Every: 1}}); err == nil {
		t.Fatal("weekly rule without weekdays accepted")
	}
	v, err := NewCreateTodo(d).Execute(ctx, TodoInput{Title: "Bins out", Due: due,
		Repeat: &RepeatInput{Kind: "weekly", Every: 1, Days: []int{4}}, Reminder: &ReminderInput{At: "19:00"}})
	if err != nil || v.Repeat.Kind != "weekly" || v.Reminder.At != "19:00" {
		t.Fatalf("create %+v %v", v, err)
	}

	// 19:00 local has passed: the reminder is sent once.
	if n, err := NewSendDueReminders(d).Execute(ctx); err != nil || n != 1 || alerts.got[0].Kind != "reminder" || alerts.got[0].Title != "Bins out" {
		t.Fatalf("reminders %d %v %+v", n, err, alerts.got)
	}
	if n, _ := NewSendDueReminders(d).Execute(ctx); n != 0 {
		t.Fatal("reminder sent twice")
	}

	// Older clients send no schedule: it is kept.
	if v, err := NewUpdateTodo(d).Execute(ctx, v.ID, TodoInput{Title: "Bins out!", Due: due}, false); err != nil || v.Repeat.Kind != "weekly" {
		t.Fatalf("update kept schedule %+v %v", v, err)
	}

	// Completing creates the next occurrence one week later.
	if _, err := NewUpdateTodo(d).Execute(ctx, v.ID, TodoInput{Title: "Bins out!", Due: due}, true); err != nil {
		t.Fatal(err)
	}
	open, _ := NewListTodos(d).Execute(ctx, false)
	if len(open) != 1 || open[0].ID == v.ID || !open[0].Due.Equal(due.AddDate(0, 0, 7)) || open[0].Reminder.At != "19:00" {
		t.Fatalf("next occurrence %+v", open)
	}
	// Saving an already completed to-do again does not create another one.
	if _, err := NewUpdateTodo(d).Execute(ctx, v.ID, TodoInput{Title: "Bins out!", Due: due}, true); err != nil {
		t.Fatal(err)
	}
	if all, _ := NewListTodos(d).Execute(ctx, true); len(all) != 2 {
		t.Fatalf("expected 2 to-dos, got %d", len(all))
	}

	// Clearing the schedule with empty values.
	if v, err := NewUpdateTodo(d).Execute(ctx, open[0].ID, TodoInput{Title: "Bins out!", Due: open[0].Due,
		Repeat: &RepeatInput{}, Reminder: &ReminderInput{}}, false); err != nil || v.Repeat.Kind != "" || v.Reminder.At != "" {
		t.Fatalf("clear %+v %v", v, err)
	}

	// Without an alert sink nothing is sent.
	d.Alerts = nil
	if n, err := NewSendDueReminders(d).Execute(ctx); n != 0 || err != nil {
		t.Fatal("no sink must send nothing")
	}
}

// TestOfflineReplay: creates with a client ID are idempotent and updates
// carry the version they were made from (#80).
func TestOfflineReplay(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)}
	d := Deps{Repo: &memRepo{m: map[string]domain.Todo{}}, Clock: c, IDs: &ids{}}
	create, update := NewCreateTodo(d), NewUpdateTodo(d)

	first, err := create.Execute(ctx, TodoInput{Title: "Brot", ID: "client-0001"})
	if err != nil || first.ID != "client-0001" || first.Version != 1 {
		t.Fatalf("%+v %v", first, err)
	}
	again, err := create.Execute(ctx, TodoInput{Title: "Brot", ID: "client-0001"})
	if err != nil || again.ID != first.ID {
		t.Fatalf("replay: %+v %v", again, err)
	}
	if list, _ := NewListTodos(d).Execute(ctx, true); len(list) != 1 {
		t.Fatalf("duplicate: %d", len(list))
	}
	if _, err := create.Execute(ctx, TodoInput{Title: "x", ID: "../x"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("bad id: %v", err)
	}

	v2, err := update.Execute(ctx, first.ID, TodoInput{Title: "Brot kaufen", Version: 1}, false)
	if err != nil || v2.Version != 2 {
		t.Fatalf("%+v %v", v2, err)
	}
	// An offline change made on version 1 now conflicts.
	if _, err := update.Execute(ctx, first.ID, TodoInput{Title: "Brötchen", Version: 1}, true); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("conflict: %v", err)
	}
	// Without a version (older clients) the change is applied.
	if v3, err := update.Execute(ctx, first.ID, TodoInput{Title: "Brötchen"}, true); err != nil || !v3.Done || v3.Version != 3 {
		t.Fatalf("%+v %v", v3, err)
	}
}
