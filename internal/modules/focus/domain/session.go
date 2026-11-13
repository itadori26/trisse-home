package domain

import (
	"time"

	"trisse-home/internal/shared/kernel"
)

// Limits of a recorded session.
const (
	MaxMinutes = 180
	// MaxAge is how far back a finished session may be recorded (a tab
	// that was asleep reports late, but not days later).
	MaxAge = 24 * time.Hour
	// WeekDays is the number of days in the statistics.
	WeekDays = 7
)

// Session is a finished focus (work) phase of the timer (#58). The running
// timer lives in the browser; only finished sessions are recorded.
type Session struct {
	ID      string
	Minutes int
	EndedAt time.Time
}

// NewSession validates a finished session reported at now.
func NewSession(id string, minutes int, endedAt, now time.Time) (Session, error) {
	if minutes < 1 || minutes > MaxMinutes {
		return Session{}, kernel.NewValidationError("minutes", "A session lasts between 1 and 180 minutes.")
	}
	// A minute of tolerance for clocks that are slightly ahead.
	if endedAt.After(now.Add(time.Minute)) || endedAt.Before(now.Add(-MaxAge)) {
		return Session{}, kernel.NewValidationError("endedAt", "The session must have ended within the last day.")
	}
	return Session{ID: id, Minutes: minutes, EndedAt: endedAt}, nil
}

// Day is the statistic of one calendar day.
type Day struct {
	Date     time.Time // midnight in the user's time zone
	Sessions int
	Minutes  int
}

// WeekStart is the first day of the statistics ending today.
func WeekStart(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d-(WeekDays-1), 0, 0, 0, 0, now.Location())
}

// Week counts sessions per day for the last seven days including today,
// oldest first, in the time zone of now.
func Week(sessions []Session, now time.Time) []Day {
	start := WeekStart(now)
	days := make([]Day, WeekDays)
	for i := range days {
		days[i].Date = start.AddDate(0, 0, i)
	}
	for _, s := range sessions {
		t := s.EndedAt.In(now.Location())
		y, m, d := t.Date()
		day := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
		for i := range days {
			if days[i].Date.Equal(day) {
				days[i].Sessions++
				days[i].Minutes += s.Minutes
			}
		}
	}
	return days
}
