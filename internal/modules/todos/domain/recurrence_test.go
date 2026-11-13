package domain

import (
	"testing"
	"time"
)

func day(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustRepeat(t *testing.T, kind RepeatKind, every int, days ...int) Recurrence {
	t.Helper()
	r, err := NewRecurrence(kind, every, days)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRecurrenceNext(t *testing.T) {
	// 2026-11-12 is a Thursday.
	cases := []struct {
		name        string
		rule        Recurrence
		due, doneOn string
		want        string
	}{
		{"daily on time", mustRepeat(t, RepeatDaily, 1), "2026-11-12", "2026-11-12", "2026-11-13"},
		{"daily done early", mustRepeat(t, RepeatDaily, 1), "2026-11-12", "2026-11-10", "2026-11-13"},
		{"overdue daily skips past days", mustRepeat(t, RepeatDaily, 1), "2026-11-09", "2026-11-12", "2026-11-13"},
		{"every 3 days", mustRepeat(t, RepeatDaily, 3), "2026-11-12", "2026-11-12", "2026-11-15"},
		{"weekly Mon+Thu", mustRepeat(t, RepeatWeekly, 1, 1, 4), "2026-11-12", "2026-11-12", "2026-11-16"},
		{"weekly Mon+Thu from Monday", mustRepeat(t, RepeatWeekly, 1, 1, 4), "2026-11-16", "2026-11-16", "2026-11-19"},
		{"every 2 weeks on Thursday", mustRepeat(t, RepeatWeekly, 2, 4), "2026-11-12", "2026-11-12", "2026-11-26"},
		{"monthly", mustRepeat(t, RepeatMonthly, 1), "2026-11-12", "2026-11-12", "2026-12-12"},
		{"monthly from the 31st clamps", mustRepeat(t, RepeatMonthly, 1), "2027-01-31", "2027-01-31", "2027-02-28"},
		{"monthly keeps the 31st after a short month", mustRepeat(t, RepeatMonthly, 1), "2027-01-31", "2027-03-01", "2027-03-31"},
		{"quarterly", mustRepeat(t, RepeatMonthly, 3), "2026-11-12", "2026-11-12", "2027-02-12"},
		{"after completion", mustRepeat(t, RepeatAfter, 10), "2026-11-01", "2026-11-12", "2026-11-22"},
		{"no due date counts from completion", mustRepeat(t, RepeatDaily, 2), "", "2026-11-12", "2026-11-14"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var due time.Time
			if tc.due != "" {
				due = day(tc.due)
			}
			if got := tc.rule.Next(due, day(tc.doneOn)); !got.Equal(day(tc.want)) {
				t.Fatalf("got %s, want %s", got.Format(time.DateOnly), tc.want)
			}
		})
	}
}

func TestNewRecurrenceValidates(t *testing.T) {
	for _, bad := range []struct {
		kind  RepeatKind
		every int
		days  []int
	}{
		{"hourly", 1, nil},
		{RepeatDaily, 0, nil},
		{RepeatDaily, 366, nil},
		{RepeatWeekly, 1, nil},
		{RepeatWeekly, 1, []int{7}},
	} {
		if _, err := NewRecurrence(bad.kind, bad.every, bad.days); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	r := mustRepeat(t, RepeatWeekly, 1, 4, 1, 4)
	if d := r.Days(); len(d) != 2 || d[0] != 1 || d[1] != 4 {
		t.Fatalf("days %v", d)
	}
	if none, _ := NewRecurrence(RepeatNone, 0, nil); none.Repeats() {
		t.Fatal("none repeats")
	}
}

func TestReminder(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	r, err := NewReminder("08:30", 1)
	if err != nil || r.At() != "08:30" {
		t.Fatalf("%v %v", r.At(), err)
	}
	if got := r.FireAt(day("2026-11-12"), loc); !got.Equal(time.Date(2026, 11, 11, 8, 30, 0, 0, loc)) {
		t.Fatalf("fire at %v", got)
	}
	for _, bad := range []struct {
		at   string
		days int
	}{{"8:30", 0}, {"25:00", 0}, {"08:00", -1}, {"08:00", 31}} {
		if _, err := NewReminder(bad.at, bad.days); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}

	todo, _ := NewTodo("t1", "Pay rent", day("2026-11-12"), 0, time.Now())
	if err := todo.Schedule(Recurrence{}, r); err != nil {
		t.Fatal(err)
	}
	before := time.Date(2026, 11, 11, 8, 29, 0, 0, loc)
	after := time.Date(2026, 11, 11, 8, 30, 0, 0, loc)
	if todo.ReminderDue(before, loc) || !todo.ReminderDue(after, loc) {
		t.Fatal("reminder must fire at 08:30 the day before")
	}
	todo.MarkReminded()
	if todo.ReminderDue(after.Add(time.Hour), loc) {
		t.Fatal("reminders are sent once per due date")
	}
	_ = todo.Edit("Pay rent", day("2026-11-20"))
	if todo.ReminderDue(after, loc) || !todo.ReminderDue(time.Date(2026, 11, 19, 9, 0, 0, 0, loc), loc) {
		t.Fatal("a new due date re-arms the reminder")
	}
	todo.SetDone(true, after)
	if todo.ReminderDue(time.Date(2026, 11, 19, 9, 0, 0, 0, loc), loc) {
		t.Fatal("done to-dos are not reminded")
	}

	undated, _ := NewTodo("t2", "Someday", time.Time{}, 0, time.Now())
	if err := undated.Schedule(Recurrence{}, r); err == nil {
		t.Fatal("reminder without due date accepted")
	}
}

func TestNextOccurrence(t *testing.T) {
	r, _ := NewReminder("09:00", 0)
	todo, _ := NewTodo("t1", "Water plants", day("2026-11-12"), 3, time.Now())
	_ = todo.Schedule(mustRepeat(t, RepeatWeekly, 1, 4), r)
	todo.MarkReminded()
	next, ok := todo.NextOccurrence("t2", 7, day("2026-11-12"), time.Now())
	if !ok || next.ID() != "t2" || next.Title() != "Water plants" || !next.Due().Equal(day("2026-11-19")) ||
		next.Position() != 7 || next.Done() || !next.RemindedFor().IsZero() || next.Reminder().At() != "09:00" {
		t.Fatalf("next %+v", next)
	}
	plain, _ := NewTodo("t3", "Once", time.Time{}, 0, time.Now())
	if _, ok := plain.NextOccurrence("t4", 0, day("2026-11-12"), time.Now()); ok {
		t.Fatal("one-off to-dos do not repeat")
	}
}
