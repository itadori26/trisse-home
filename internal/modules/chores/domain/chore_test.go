package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

var now = time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)

func TestNew(t *testing.T) {
	c, err := New("c1", Input{Title: " Take out the bins ", EveryDays: 7, Due: "2026-11-23", Assignees: []string{"Anna", " ", "Ben "}}, kernel.Access{}, now)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Take out the bins" || !slices.Equal(c.Assignees, []string{"Anna", "Ben"}) || c.Assignee() != "Anna" || c.DueDate() != "2026-11-23" {
		t.Fatalf("%+v", c)
	}
	for name, in := range map[string]Input{
		"no title":   {EveryDays: 1, Due: "2026-11-23"},
		"long title": {Title: strings.Repeat("x", 81), EveryDays: 1, Due: "2026-11-23"},
		"never":      {Title: "x", Due: "2026-11-23"},
		"too rare":   {Title: "x", EveryDays: 366, Due: "2026-11-23"},
		"no due":     {Title: "x", EveryDays: 1},
		"long name":  {Title: "x", EveryDays: 1, Due: "2026-11-23", Assignees: []string{strings.Repeat("x", 41)}},
		"crowd":      {Title: "x", EveryDays: 1, Due: "2026-11-23", Assignees: slices.Repeat([]string{"x"}, 9)},
	} {
		if _, err := New("c", in, kernel.Access{}, now); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestDoneRotatesAndMovesOn(t *testing.T) {
	c, _ := New("c1", Input{Title: "Dishes", EveryDays: 2, Due: "2026-11-23", Assignees: []string{"Anna", "Ben", "Carla"}}, kernel.Access{}, now)
	today := now
	if !c.DueOn(today) || c.OverdueOn(today) {
		t.Fatal("due today, not overdue")
	}
	c = c.Done(today, now)
	if c.DueDate() != "2026-11-25" || c.Assignee() != "Ben" || c.LastDoneBy != "Anna" || !c.LastDoneAt.Equal(now) || c.DueOn(today) {
		t.Fatalf("after one round: %+v", c)
	}
	// Done early: the round still counts from the due day.
	c = c.Done(today, now)
	if c.DueDate() != "2026-11-27" || c.Assignee() != "Carla" {
		t.Fatalf("early: %+v", c)
	}
	// Overdue by a week: the next round counts from today.
	late := time.Date(2026, 12, 4, 9, 0, 0, 0, time.UTC)
	if !c.OverdueOn(late) {
		t.Fatal("should be overdue")
	}
	c = c.Done(late, late)
	if c.DueDate() != "2026-12-06" || c.Assignee() != "Anna" {
		t.Fatalf("late: %+v", c)
	}
}

func TestUpdateKeepsTurnInRange(t *testing.T) {
	c, _ := New("c1", Input{Title: "Dishes", EveryDays: 1, Due: "2026-11-23", Assignees: []string{"Anna", "Ben", "Carla"}}, kernel.Access{}, now)
	c = c.Done(now, now).Done(now, now) // Carla's turn
	c, err := c.Update(Input{Title: "Dishes", EveryDays: 1, Due: "2026-11-25", Assignees: []string{"Anna", "Ben"}})
	if err != nil || c.Turn != 0 || c.Assignee() != "Anna" {
		t.Fatalf("%+v %v", c, err)
	}
	c, _ = c.Update(Input{Title: "Dishes", EveryDays: 1, Due: "2026-11-25"})
	if c.Assignee() != "" || c.Done(now, now).LastDoneBy != "" {
		t.Fatalf("no assignees: %+v", c)
	}
}
