package domain

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of a habit.
const (
	MaxNameLength  = 60
	MaxEmojiLength = 8
	// HeatmapDays is the length of the heatmap (four weeks).
	HeatmapDays = 28
)

// ErrNotFound is returned for unknown habits.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The habit does not exist.")

// ErrDayNotEditable is returned for days other than today and yesterday.
var ErrDayNotEditable = kernel.NewValidationError("date", "Only today and yesterday can be checked off.")

// Schedule says on which weekdays a habit is due; empty means every day.
type Schedule struct{ days []time.Weekday }

// NewSchedule validates weekdays 0 (Sunday) to 6; duplicates are dropped.
func NewSchedule(days []int) (Schedule, error) {
	var s Schedule
	for _, d := range days {
		if d < 0 || d > 6 {
			return Schedule{}, kernel.NewValidationError("days", "Weekdays are 0 (Sunday) to 6.")
		}
		if !slices.Contains(s.days, time.Weekday(d)) {
			s.days = append(s.days, time.Weekday(d))
		}
	}
	slices.Sort(s.days)
	// All seven days are the same as "every day".
	if len(s.days) == 7 {
		s.days = nil
	}
	return s, nil
}

// Daily reports whether the habit is due every day.
func (s Schedule) Daily() bool { return len(s.days) == 0 }

// Due reports whether the habit is due on the day.
func (s Schedule) Due(day time.Time) bool { return s.Daily() || slices.Contains(s.days, day.Weekday()) }

// Days returns the weekdays as numbers (empty = every day).
func (s Schedule) Days() []int {
	out := make([]int, 0, len(s.days))
	for _, d := range s.days {
		out = append(out, int(d))
	}
	return out
}

// Habit is a small daily habit to keep (#106).
type Habit struct {
	ID        string
	Name      string
	Emoji     string
	Schedule  Schedule
	Position  int
	CreatedAt time.Time
}

// New validates and creates a habit.
func New(id, name, emoji string, schedule Schedule, position int, now time.Time) (Habit, error) {
	h := Habit{ID: id, Position: position, CreatedAt: now}
	return h.Update(name, emoji, schedule)
}

// Update changes name, emoji and schedule.
func (h Habit) Update(name, emoji string, schedule Schedule) (Habit, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return Habit{}, kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	emoji = strings.TrimSpace(emoji)
	if utf8.RuneCountInString(emoji) > MaxEmojiLength {
		return Habit{}, kernel.NewValidationError("emoji", "Use a single emoji.")
	}
	h.Name, h.Emoji, h.Schedule = name, emoji, schedule
	return h, nil
}

// Day is a calendar day (midnight UTC); only the date counts.
type Day = time.Time

// DayOf returns the calendar day of a moment in its own time zone.
func DayOf(t time.Time) Day {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// EditableDay checks that a day is today or yesterday.
func EditableDay(day, today Day) error {
	if day.Equal(today) || day.Equal(today.AddDate(0, 0, -1)) {
		return nil
	}
	return ErrDayNotEditable
}

// HeatCell is one day of the heatmap.
type HeatCell struct {
	Day  Day
	Due  bool
	Done bool
}

// Stats are computed at read time from the completions (single source of truth).
type Stats struct {
	DoneToday bool
	Current   int
	Best      int
	Heatmap   []HeatCell // oldest first, ends today
}

// ComputeStats derives streaks and the heatmap. A streak counts due days
// in a row that were done; days the habit is not due neither count nor
// break it, and today does not break the streak while it is still open.
func ComputeStats(h Habit, done map[Day]bool, today Day) Stats {
	st := Stats{DoneToday: done[today]}

	// Current streak: walk back from today (or yesterday, if today is still open).
	day := today
	if h.Schedule.Due(today) && !done[today] {
		day = today.AddDate(0, 0, -1)
	}
	start := DayOf(h.CreatedAt)
	for earliest := firstDone(done, start); !day.Before(earliest); day = day.AddDate(0, 0, -1) {
		if !h.Schedule.Due(day) {
			continue
		}
		if !done[day] {
			break
		}
		st.Current++
	}

	// Best streak: one pass from the first completion to today.
	run := 0
	for d := firstDone(done, start); !d.After(today); d = d.AddDate(0, 0, 1) {
		if !h.Schedule.Due(d) {
			continue
		}
		switch {
		case done[d]:
			run++
			st.Best = max(st.Best, run)
		case d.Equal(today):
			// still open
		default:
			run = 0
		}
	}

	for i := HeatmapDays - 1; i >= 0; i-- {
		d := today.AddDate(0, 0, -i)
		st.Heatmap = append(st.Heatmap, HeatCell{Day: d, Due: h.Schedule.Due(d), Done: done[d]})
	}
	return st
}

// firstDone is the earliest completed day, or start when there is none
// or an older completion was imported.
func firstDone(done map[Day]bool, start Day) Day {
	first := start
	for d, ok := range done {
		if ok && d.Before(first) {
			first = d
		}
	}
	return first
}
