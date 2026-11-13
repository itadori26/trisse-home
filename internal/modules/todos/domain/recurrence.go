package domain

import (
	"slices"
	"time"

	"trisse-home/internal/shared/kernel"
)

// RepeatKind selects how a to-do recurs (#68).
type RepeatKind string

// Recurrence kinds. Fixed schedules follow the due date; "after" counts
// from the day the to-do was completed.
const (
	RepeatNone    RepeatKind = ""
	RepeatDaily   RepeatKind = "daily"
	RepeatWeekly  RepeatKind = "weekly"
	RepeatMonthly RepeatKind = "monthly"
	RepeatAfter   RepeatKind = "after"
)

// MaxRepeatEvery limits the interval ("every 365 days").
const MaxRepeatEvery = 365

// Recurrence is a value object: kind, interval and – for weekly – weekdays.
type Recurrence struct {
	kind  RepeatKind
	every int
	days  []time.Weekday
}

var errRecurrence = kernel.NewValidationError("repeat", "Repeat every 1–365 days, weeks or months; weekly repeats need at least one weekday.")

// NewRecurrence validates a rule. Weekdays are only used by weekly rules.
func NewRecurrence(kind RepeatKind, every int, days []int) (Recurrence, error) {
	if kind == RepeatNone {
		return Recurrence{}, nil
	}
	if !slices.Contains([]RepeatKind{RepeatDaily, RepeatWeekly, RepeatMonthly, RepeatAfter}, kind) || every < 1 || every > MaxRepeatEvery {
		return Recurrence{}, errRecurrence
	}
	r := Recurrence{kind: kind, every: every}
	if kind == RepeatWeekly {
		for _, d := range days {
			if d < 0 || d > 6 {
				return Recurrence{}, errRecurrence
			}
			if !slices.Contains(r.days, time.Weekday(d)) {
				r.days = append(r.days, time.Weekday(d))
			}
		}
		if len(r.days) == 0 {
			return Recurrence{}, errRecurrence
		}
		slices.Sort(r.days)
	}
	return r, nil
}

// Kind returns the recurrence kind (RepeatNone if the to-do does not repeat).
func (r Recurrence) Kind() RepeatKind { return r.kind }

// Every returns the interval.
func (r Recurrence) Every() int { return r.every }

// Days returns the weekdays of a weekly rule (0 = Sunday).
func (r Recurrence) Days() []int {
	out := make([]int, 0, len(r.days))
	for _, d := range r.days {
		out = append(out, int(d))
	}
	return out
}

// Repeats reports whether the rule creates further occurrences.
func (r Recurrence) Repeats() bool { return r.kind != RepeatNone }

func dateOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// addMonths keeps the day of month and clamps to the month's last day
// (31 Jan + 1 month = 28/29 Feb).
func addMonths(d time.Time, n, day int) time.Time {
	first := time.Date(d.Year(), d.Month()+time.Month(n), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(day, last)-1)
}

// Next returns the due date of the next occurrence. due is the current due
// date (zero = none) and doneOn the day the to-do was completed. Fixed
// schedules continue from the due date but never land on or before doneOn,
// so an overdue daily task is not recreated for a day already past.
func (r Recurrence) Next(due, doneOn time.Time) time.Time {
	doneOn = dateOf(doneOn)
	if r.kind == RepeatNone {
		return time.Time{}
	}
	if r.kind == RepeatAfter {
		return doneOn.AddDate(0, 0, r.every)
	}
	base := doneOn
	if !due.IsZero() {
		base = dateOf(due)
	}
	day := base.Day()
	next := base
	for i := 0; i < 5000; i++ { // bounded: at most ~14 years of daily steps
		next = r.step(next, base, day)
		if next.After(doneOn) {
			return next
		}
	}
	return doneOn.AddDate(0, 0, 1)
}

// step returns the occurrence after d. Weekly rules with an interval > 1
// count weeks from the week of base.
func (r Recurrence) step(d, base time.Time, day int) time.Time {
	switch r.kind {
	case RepeatDaily:
		return d.AddDate(0, 0, r.every)
	case RepeatMonthly:
		months := (d.Year()-base.Year())*12 + int(d.Month()-base.Month()) + r.every
		return addMonths(base, months, day)
	default: // weekly
		weekStart := func(t time.Time) time.Time { return t.AddDate(0, 0, -int(t.Weekday())) }
		for c := d.AddDate(0, 0, 1); ; c = c.AddDate(0, 0, 1) {
			weeks := int(weekStart(c).Sub(weekStart(base)).Hours() / 24 / 7)
			if weeks%r.every == 0 && slices.Contains(r.days, c.Weekday()) {
				return c
			}
		}
	}
}

// Reminder sends one push on the due date (or some days before) at a time
// of day (#68).
type Reminder struct {
	set        bool
	daysBefore int
	minute     int // minutes since midnight
}

// MaxReminderDaysBefore limits how early a reminder can be.
const MaxReminderDaysBefore = 30

// NewReminder validates a reminder; at is "HH:MM", empty means none.
func NewReminder(at string, daysBefore int) (Reminder, error) {
	if at == "" {
		return Reminder{}, nil
	}
	t, err := time.Parse("15:04", at)
	if err != nil || len(at) != len("15:04") || daysBefore < 0 || daysBefore > MaxReminderDaysBefore {
		return Reminder{}, kernel.NewValidationError("reminder", "Reminders need a time like 08:00 and at most 30 days before the due date.")
	}
	return Reminder{set: true, daysBefore: daysBefore, minute: t.Hour()*60 + t.Minute()}, nil
}

// Set reports whether a reminder is configured.
func (r Reminder) Set() bool { return r.set }

// At returns the time of day as "HH:MM" ("" if none).
func (r Reminder) At() string {
	if !r.set {
		return ""
	}
	return time.Date(0, 1, 1, r.minute/60, r.minute%60, 0, 0, time.UTC).Format("15:04")
}

// DaysBefore returns how many days before the due date it fires.
func (r Reminder) DaysBefore() int { return r.daysBefore }

// FireAt returns when the reminder for a due date fires in loc.
func (r Reminder) FireAt(due time.Time, loc *time.Location) time.Time {
	y, m, d := due.AddDate(0, 0, -r.daysBefore).Date()
	return time.Date(y, m, d, r.minute/60, r.minute%60, 0, 0, loc)
}
