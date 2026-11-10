package domain

import (
	"testing"
	"time"
)

func TestTodo(t *testing.T) {
	now := time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)
	if _, err := NewTodo("1", "  ", time.Time{}, 0, now); err == nil {
		t.Fatal("expected error for blank title")
	}
	due := time.Date(2026, 11, 9, 23, 30, 0, 0, time.UTC)
	todo, err := NewTodo("1", "Pay rent", due, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if todo.Due().Hour() != 0 {
		t.Fatal("due date not truncated to the day")
	}
	if !todo.IsOverdue(now) {
		t.Fatal("expected overdue")
	}
	todo.SetDone(true, now)
	if !todo.Done() || todo.IsOverdue(now) {
		t.Fatal("done to-dos are never overdue")
	}
	todo.SetDone(true, now.Add(time.Hour))
	if !todo.CompletedAt().Equal(now) {
		t.Fatal("completion time must not change when marking done twice")
	}
	todo.SetDone(false, now)
	if todo.Done() {
		t.Fatal("reopen failed")
	}
}
