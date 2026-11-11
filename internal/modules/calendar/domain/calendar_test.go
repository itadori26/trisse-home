package domain

import (
	"testing"
	"time"
)

func TestNewSettings(t *testing.T) {
	s, err := NewSettings([]string{" https://cloud.example.com/remote.php/dav/calendars/me/personal ", "", "https://cloud.example.com/remote.php/dav/calendars/me/personal/"}, "me", "app-pw")
	if err != nil || len(s.CalendarURLs) != 1 || s.CalendarURLs[0][len(s.CalendarURLs[0])-1] != '/' {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	if _, err := NewSettings(nil, "me", "pw"); err == nil {
		t.Fatal("expected error without urls")
	}
	if _, err := NewSettings([]string{"ftp://x"}, "me", "pw"); err == nil {
		t.Fatal("expected error for bad scheme")
	}
}

func TestCalendarName(t *testing.T) {
	if got := CalendarName("https://c.example/remote.php/dav/calendars/me/work%20stuff/"); got != "work stuff" {
		t.Fatalf("got %q", got)
	}
}

func TestSortAndEndsAfter(t *testing.T) {
	day := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	events := []Event{
		{Summary: "late", Start: day.Add(18 * time.Hour), End: day.Add(19 * time.Hour)},
		{Summary: "allday", Start: day, End: day.AddDate(0, 0, 1), AllDay: true},
		{Summary: "early", Start: day.Add(8 * time.Hour), End: day.Add(9 * time.Hour)},
	}
	SortEvents(events)
	if events[0].Summary != "allday" || events[1].Summary != "early" {
		t.Fatalf("order %v", events)
	}
	now := day.Add(10 * time.Hour)
	if events[1].EndsAfter(now) || !events[0].EndsAfter(now) || !events[2].EndsAfter(now) {
		t.Fatal("EndsAfter wrong")
	}
}

func TestNewEventAndRange(t *testing.T) {
	day := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	if _, err := NewEvent("u", " ", "", day, day.Add(time.Hour), false); err == nil {
		t.Fatal("empty title accepted")
	}
	if _, err := NewEvent("u", "x", "", day, day, false); err == nil {
		t.Fatal("zero length accepted")
	}
	e, err := NewEvent("u", " Dentist ", " Main St ", day.Add(9*time.Hour), day.Add(10*time.Hour), false)
	if err != nil || e.Summary != "Dentist" || e.Location != "Main St" {
		t.Fatalf("%+v %v", e, err)
	}
	if e, err := NewEvent("u", "Trip", "", day.Add(5*time.Hour), day.AddDate(0, 0, 2), true); err != nil || !e.Start.Equal(day) {
		t.Fatalf("all-day %+v %v", e, err)
	}
	if CheckRange(day, day.AddDate(0, 0, 42)) != nil || CheckRange(day, day.AddDate(0, 0, 43)) == nil || CheckRange(day, day) == nil {
		t.Fatal("range check wrong")
	}
	s := Settings{CalendarURLs: []string{"https://c/cal/home/"}}
	if !s.Contains("https://c/cal/home") || s.Contains("https://c/cal/other/") {
		t.Fatal("contains wrong")
	}
}
