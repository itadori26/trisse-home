package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of a countdown.
const (
	MaxTitleLength = 120
	MaxEmojiLength = 8 // runes; flags and skin tones take several
	// MaxGraceDays bounds how long a past countdown stays visible.
	MaxGraceDays = 365
)

// ErrNotFound is returned for unknown countdowns.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The countdown does not exist.")

// Countdown is a named point in time to count down to (#54): a calendar
// day, optionally with a time of day. Without a time it ends at the start
// of that day.
type Countdown struct {
	ID    string
	Title string
	Emoji string
	// Day is the calendar day (midnight UTC; only the date counts).
	Day time.Time
	// Minute is the time of day in minutes after midnight, or -1 for the
	// whole day.
	Minute    int
	CreatedAt time.Time
}

// Input is what a user enters for a countdown.
type Input struct {
	Title string
	Emoji string
	Date  string // YYYY-MM-DD
	Time  string // HH:MM or empty
}

// New validates the input and creates a countdown.
func New(id string, in Input, now time.Time) (Countdown, error) {
	c := Countdown{ID: id, CreatedAt: now}
	if err := c.apply(in); err != nil {
		return Countdown{}, err
	}
	return c, nil
}

// Update changes title, emoji, day and time.
func (c Countdown) Update(in Input) (Countdown, error) {
	err := c.apply(in)
	return c, err
}

func (c *Countdown) apply(in Input) error {
	title := strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLength {
		return kernel.NewValidationError("title", "The title must be between 1 and 120 characters long.")
	}
	emoji := strings.TrimSpace(in.Emoji)
	if utf8.RuneCountInString(emoji) > MaxEmojiLength {
		return kernel.NewValidationError("emoji", "Use a single emoji.")
	}
	day, err := time.Parse(time.DateOnly, strings.TrimSpace(in.Date))
	if err != nil {
		return kernel.NewValidationError("date", "Enter a date like 2026-12-24.")
	}
	minute := -1
	if t := strings.TrimSpace(in.Time); t != "" {
		tod, err := time.Parse("15:04", t)
		if err != nil {
			return kernel.NewValidationError("time", "Enter a time like 07:00.")
		}
		minute = tod.Hour()*60 + tod.Minute()
	}
	c.Title, c.Emoji, c.Day, c.Minute = title, emoji, day, minute
	return nil
}

// AllDay reports whether the countdown has no time of day.
func (c Countdown) AllDay() bool { return c.Minute < 0 }

// Date is the day as YYYY-MM-DD.
func (c Countdown) Date() string { return c.Day.Format(time.DateOnly) }

// Time is the time of day as HH:MM, or empty for the whole day.
func (c Countdown) Time() string {
	if c.AllDay() {
		return ""
	}
	return time.Date(0, 1, 1, c.Minute/60, c.Minute%60, 0, 0, time.UTC).Format("15:04")
}

// TargetIn is the moment the countdown ends in the given time zone.
func (c Countdown) TargetIn(loc *time.Location) time.Time {
	y, m, d := c.Day.Date()
	t := time.Date(y, m, d, 0, 0, 0, 0, loc)
	if !c.AllDay() {
		t = t.Add(time.Duration(c.Minute) * time.Minute)
	}
	return t
}

// IsPast reports whether the countdown has ended at now. An all-day
// countdown counts as running until the end of its day.
func (c Countdown) IsPast(now time.Time) bool {
	end := c.TargetIn(now.Location())
	if c.AllDay() {
		end = end.AddDate(0, 0, 1)
	}
	return !now.Before(end)
}

// Hidden reports whether a past countdown has outlived the grace period
// of graceDays full days after it ended.
func (c Countdown) Hidden(now time.Time, graceDays int) bool {
	if !c.IsPast(now) {
		return false
	}
	end := c.TargetIn(now.Location())
	if c.AllDay() {
		end = end.AddDate(0, 0, 1)
	}
	return !now.Before(end.AddDate(0, 0, graceDays))
}

// ClampGrace keeps a requested grace period within 0..MaxGraceDays.
func ClampGrace(days int) int { return min(max(days, 0), MaxGraceDays) }
