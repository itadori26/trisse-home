package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

var berlin = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	return loc
}()

func TestNewValidates(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	cases := map[string]Input{
		"empty title":  {Title: " ", Date: "2026-12-24"},
		"long title":   {Title: strings.Repeat("x", 121), Date: "2026-12-24"},
		"bad date":     {Title: "X", Date: "24.12.2026"},
		"bad time":     {Title: "X", Date: "2026-12-24", Time: "25:00"},
		"emoji string": {Title: "X", Date: "2026-12-24", Emoji: "not an emoji"},
	}
	for name, in := range cases {
		if _, err := New("1", in, now); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	c, err := New("1", Input{Title: " Exam ", Emoji: "📚", Date: "2026-12-24", Time: "09:30"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Exam" || c.Date() != "2026-12-24" || c.Time() != "09:30" || c.AllDay() {
		t.Fatalf("%+v", c)
	}
	whole, _ := New("2", Input{Title: "Holiday", Date: "2026-12-24"}, now)
	if !whole.AllDay() || whole.Time() != "" {
		t.Fatalf("%+v", whole)
	}
}

func TestIsPastAndHidden(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	allDay, _ := New("1", Input{Title: "Day", Date: "2026-12-24"}, now)
	timed, _ := New("2", Input{Title: "Exam", Date: "2026-12-24", Time: "09:30"}, now)
	at := func(day, hour, minute int) time.Time { return time.Date(2026, 12, day, hour, minute, 0, 0, berlin) }

	if got := timed.TargetIn(berlin); !got.Equal(at(24, 9, 30)) {
		t.Fatalf("target %v", got)
	}
	// A whole-day countdown runs until the end of its day.
	if allDay.IsPast(at(24, 23, 59)) || !allDay.IsPast(at(25, 0, 0)) {
		t.Error("all-day end")
	}
	if timed.IsPast(at(24, 9, 29)) || !timed.IsPast(at(24, 9, 30)) {
		t.Error("timed end")
	}
	// Grace period in whole days after the end.
	if timed.Hidden(at(25, 9, 29), 1) || !timed.Hidden(at(25, 9, 30), 1) {
		t.Error("grace 1 day")
	}
	if !timed.Hidden(at(24, 9, 30), 0) {
		t.Error("grace 0 hides at once")
	}
	if allDay.Hidden(at(1, 0, 0), 0) {
		t.Error("future countdown hidden")
	}
}

func TestClampGrace(t *testing.T) {
	if ClampGrace(-3) != 0 || ClampGrace(5) != 5 || ClampGrace(9999) != MaxGraceDays {
		t.Fatal("clamp")
	}
}
