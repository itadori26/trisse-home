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
