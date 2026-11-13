package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestNewSession(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	for name, c := range map[string]struct {
		minutes int
		ended   time.Time
	}{
		"zero minutes":  {0, now},
		"too long":      {181, now},
		"in the future": {25, now.Add(2 * time.Minute)},
		"too old":       {25, now.Add(-25 * time.Hour)},
	} {
		if _, err := NewSession("1", c.minutes, c.ended, now); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := NewSession("1", 25, now.Add(30*time.Second), now); err != nil {
		t.Fatalf("small clock skew: %v", err)
	}
}

func TestWeek(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, loc)
	at := func(day, hour int) time.Time { return time.Date(2026, 11, day, hour, 0, 0, 0, loc) }
	sessions := []Session{
		{Minutes: 25, EndedAt: at(12, 9)},
		{Minutes: 25, EndedAt: at(12, 8)},
		{Minutes: 50, EndedAt: at(6, 23)},
		// 23:30 UTC on the 10th is already the 11th in CET.
		{Minutes: 25, EndedAt: time.Date(2026, 11, 10, 23, 30, 0, 0, time.UTC)},
		{Minutes: 25, EndedAt: at(5, 12)}, // outside the week
	}
	week := Week(sessions, now)
	if len(week) != WeekDays || !week[0].Date.Equal(at(6, 0)) || !week[6].Date.Equal(at(12, 0)) {
		t.Fatalf("days %+v", week)
	}
	if week[6].Sessions != 2 || week[6].Minutes != 50 || week[0].Sessions != 1 || week[5].Sessions != 1 || week[4].Sessions != 0 {
		t.Fatalf("counts %+v", week)
	}
}
