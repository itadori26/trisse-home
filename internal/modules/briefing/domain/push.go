package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// PushSchedule sends the briefing once a day as a notification (#70).
// It is off by default.
type PushSchedule struct {
	Enabled bool
	// Minute is the time of day in minutes since midnight.
	Minute int
	// Days are the weekdays it is sent on (empty = every day).
	Days []time.Weekday
	// SkipEmpty sends nothing when there is nothing to report.
	SkipEmpty bool
}

// DefaultPushSchedule is off, at 07:00 on every day, skipping empty ones.
func DefaultPushSchedule() PushSchedule {
	return PushSchedule{Minute: 7 * 60, SkipEmpty: true}
}

// NewPushSchedule validates a schedule; at is "HH:MM".
func NewPushSchedule(enabled bool, at string, days []int, skipEmpty bool) (PushSchedule, error) {
	t, err := time.Parse("15:04", at)
	if err != nil || len(at) != len("15:04") {
		return PushSchedule{}, kernel.NewValidationError("time", "Enter a time like 07:00.")
	}
	s := PushSchedule{Enabled: enabled, Minute: t.Hour()*60 + t.Minute(), SkipEmpty: skipEmpty}
	for _, d := range days {
		if d < 0 || d > 6 {
			return PushSchedule{}, kernel.NewValidationError("days", "Weekdays are 0 (Sunday) to 6.")
		}
		if !slices.Contains(s.Days, time.Weekday(d)) {
			s.Days = append(s.Days, time.Weekday(d))
		}
	}
	slices.Sort(s.Days)
	return s, nil
}

// At returns the time of day as "HH:MM".
func (s PushSchedule) At() string { return fmt.Sprintf("%02d:%02d", s.Minute/60, s.Minute%60) }

// DayNumbers returns the weekdays as numbers (0 = Sunday).
func (s PushSchedule) DayNumbers() []int {
	out := make([]int, 0, len(s.Days))
	for _, d := range s.Days {
		out = append(out, int(d))
	}
	return out
}

// Due reports whether the briefing should be sent at now (in the user's
// time zone): on an enabled day, once the time has passed, and not yet
// sent today. A server that was down at the time sends it later that day.
func (s PushSchedule) Due(now, lastSent time.Time) bool {
	if !s.Enabled || (len(s.Days) > 0 && !slices.Contains(s.Days, now.Weekday())) {
		return false
	}
	y, m, d := now.Date()
	fire := time.Date(y, m, d, s.Minute/60, s.Minute%60, 0, 0, now.Location())
	return !now.Before(fire) && lastSent.Before(fire)
}

// labels are the push headings per section; push texts are English like
// the other server notifications.
var labels = map[SectionKind]string{
	SectionWarnings:  "Weather warnings",
	SectionIncidents: "Since yesterday evening",
	SectionServices:  "Offline",
	SectionTodos:     "To-dos",
	SectionEvents:    "Appointments",
	SectionWaste:     "Waste collection",
}

// Text renders the briefing as plain text for a push notification, in
// the same order and with the same content as the widget.
func (b Briefing) Text(loc *time.Location) string {
	if b.Empty() {
		return "Nothing to report."
	}
	var sb strings.Builder
	for i, s := range b.Sections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(labels[s.Kind] + ":\n")
		for _, it := range s.Items {
			sb.WriteString("• " + it.Text)
			if !it.At.IsZero() && (s.Kind == SectionEvents || s.Kind == SectionIncidents) {
				sb.WriteString(" (" + it.At.In(loc).Format("15:04") + ")")
			}
			if it.Detail != "" {
				sb.WriteString(" – " + it.Detail)
			}
			sb.WriteString("\n")
		}
		if s.More > 0 {
			fmt.Fprintf(&sb, "• … and %d more\n", s.More)
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// MostSevere returns the highest severity in the briefing.
func (b Briefing) MostSevere() Severity {
	out := SeverityInfo
	for _, s := range b.Sections {
		for _, it := range s.Items {
			out = max(out, it.Severity)
		}
	}
	return out
}
