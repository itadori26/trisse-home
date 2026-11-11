package ical

import (
	"strings"
	"testing"
	"time"
)

func TestTodos(t *testing.T) {
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	raw := NewTodo("u1", "Milch, 2×", now)
	todos := ParseTodos(raw)
	if len(todos) != 1 || todos[0].Summary != "Milch, 2×" || todos[0].Done || !todos[0].Created.Equal(now) {
		t.Fatalf("%+v", todos)
	}
	// Other apps' properties survive and VALARM sub-components are untouched.
	foreign := strings.Replace(raw, "END:VTODO", "X-APPLE-SORT-ORDER:42\r\nPRIORITY:1\r\nBEGIN:VALARM\r\nACTION:DISPLAY\r\nEND:VALARM\r\nEND:VTODO", 1)
	done := SetTodoDone(foreign, true, now.Add(time.Hour))
	if got := ParseTodos(done); len(got) != 1 || !got[0].Done || !got[0].Completed.Equal(now.Add(time.Hour)) {
		t.Fatalf("done: %+v", got)
	}
	for _, keep := range []string{"X-APPLE-SORT-ORDER:42", "PRIORITY:1", "BEGIN:VALARM", "SUMMARY:Milch\\, 2×"} {
		if !strings.Contains(done, keep) {
			t.Fatalf("lost %q:\n%s", keep, done)
		}
	}
	if strings.Count(done, "STATUS:") != 1 {
		t.Fatalf("duplicate STATUS:\n%s", done)
	}
	undone := SetTodoDone(done, false, now)
	if got := ParseTodos(undone); got[0].Done || strings.Contains(undone, "COMPLETED:") {
		t.Fatalf("undone: %s", undone)
	}
	cancelled := strings.Replace(raw, "STATUS:NEEDS-ACTION", "STATUS:CANCELLED", 1)
	if len(ParseTodos(cancelled)) != 0 {
		t.Fatal("cancelled todo listed")
	}
}
