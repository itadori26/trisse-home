package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Severity orders notifications (#69).
type Severity string

// Severities, from least to most important.
const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Rank orders severities; unknown values count as warning.
func (s Severity) Rank() int {
	switch s {
	case SeverityInfo:
		return 0
	case SeverityCritical:
		return 2
	}
	return 1
}

// Valid reports whether s is a known severity.
func (s Severity) Valid() bool {
	return s == SeverityInfo || s == SeverityWarning || s == SeverityCritical
}

// AlertTypes are the notification types the user can configure.
var AlertTypes = []string{"service", "container", "disk", "tls", "backup", "check", "security", "diskhealth", "forecast", "speedtest", "image", "reminder", "briefing"}

// DefaultSeverities apply when the user did not choose one.
var DefaultSeverities = map[string]Severity{
	"service": SeverityCritical, "container": SeverityWarning, "disk": SeverityWarning, "tls": SeverityWarning,
	"backup": SeverityCritical, "check": SeverityCritical, "security": SeverityWarning, "diskhealth": SeverityCritical,
	"forecast": SeverityWarning, "speedtest": SeverityInfo, "image": SeverityInfo,
	"reminder": SeverityInfo, "briefing": SeverityInfo,
}

// alertTags are the ntfy tags (emoji) of one-off alerts; the security tag
// stays the fallback for unknown kinds so an unexpected alert never looks
// harmless.
var alertTags = map[string]string{"security": "lock", "reminder": "bell", "briefing": "sunny"}

// AlertTag returns the ntfy tag for a one-off alert of the given kind.
func AlertTag(kind string) string {
	if tag, ok := alertTags[kind]; ok {
		return tag
	}
	return "lock"
}

// QuietHours mute non-critical notifications during a daily period on the
// chosen weekdays; Start after End means the period crosses midnight and
// belongs to the weekday it starts on.
type QuietHours struct {
	Enabled bool
	Start   string // "22:00"
	End     string // "07:00"
	Days    []time.Weekday
}

func parseClock(s string) (int, bool) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 || len(s) != 5 {
		return 0, false
	}
	return h*60 + m, true
}

// Validate checks the times.
func (q QuietHours) Validate() error {
	if !q.Enabled {
		return nil
	}
	s, ok1 := parseClock(q.Start)
	e, ok2 := parseClock(q.End)
	if !ok1 || !ok2 || s == e {
		return kernel.NewValidationError("quietHours", "Quiet hours need a start and an end such as 22:00 and 07:00.")
	}
	return nil
}

// Active reports whether t (in its own location) is inside quiet hours.
func (q QuietHours) Active(t time.Time) bool {
	if !q.Enabled {
		return false
	}
	s, ok1 := parseClock(q.Start)
	e, ok2 := parseClock(q.End)
	if !ok1 || !ok2 || s == e {
		return false
	}
	now := t.Hour()*60 + t.Minute()
	day := t.Weekday()
	onDay := func(d time.Weekday) bool { return len(q.Days) == 0 || slices.Contains(q.Days, d) }
	if s < e {
		return onDay(day) && now >= s && now < e
	}
	// Crosses midnight: the evening part on the start day, the morning part
	// belongs to the previous day.
	if now >= s {
		return onDay(day)
	}
	return now < e && onDay((day+6)%7)
}

// Policy configures severities and quiet hours.
type Policy struct {
	Severities map[string]Severity
	Quiet      QuietHours
}

// SeverityOf returns the configured severity of an alert type.
func (p Policy) SeverityOf(alertType string) Severity {
	if s, ok := p.Severities[alertType]; ok && s.Valid() {
		return s
	}
	if s, ok := DefaultSeverities[alertType]; ok {
		return s
	}
	return SeverityWarning
}

// Decision is what happens to a notification.
type Decision int

// Decisions of the delivery policy.
const (
	Deliver Decision = iota
	// Defer collects the notification for the digest after quiet hours.
	Defer
)

// Decide applies the quiet hours: critical notifications always go out.
func (p Policy) Decide(s Severity, now time.Time) Decision {
	if s != SeverityCritical && p.Quiet.Active(now) {
		return Defer
	}
	return Deliver
}

// MaintenanceWindow mutes notifications of the scoped objects. Scope
// entries are "service:<name or id>", "container:<name>" or "*".
type MaintenanceWindow struct {
	ID    string
	Scope []string
	Start time.Time
	End   time.Time
	Note  string
}

// maxMaintenance bounds a maintenance window.
const maxMaintenance = 7 * 24 * time.Hour

// NewMaintenanceWindow validates a window.
func NewMaintenanceWindow(id string, scope []string, start, end time.Time, note string) (MaintenanceWindow, error) {
	clean := []string{}
	for _, s := range scope {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		kind, _, ok := strings.Cut(s, ":")
		if s != "*" && (!ok || (kind != "service" && kind != "container")) {
			return MaintenanceWindow{}, kernel.NewValidationErrorf("scope", "Invalid scope {scope}: use service:<name>, container:<name> or *.", "scope", s)
		}
		clean = append(clean, s)
	}
	if len(clean) == 0 {
		return MaintenanceWindow{}, kernel.NewValidationError("scope", "Choose what the maintenance applies to.")
	}
	if !end.After(start) || end.Sub(start) > maxMaintenance {
		return MaintenanceWindow{}, kernel.NewValidationError("end", "A maintenance window must end after it starts and last at most 7 days.")
	}
	if len(note) > 200 {
		note = note[:200]
	}
	return MaintenanceWindow{ID: id, Scope: clean, Start: start, End: end, Note: strings.TrimSpace(note)}, nil
}

// ActiveAt reports whether the window applies at t.
func (w MaintenanceWindow) ActiveAt(t time.Time) bool { return !t.Before(w.Start) && t.Before(w.End) }

// Covers reports whether the window mutes an observation (by key or label).
func (w MaintenanceWindow) Covers(kind, key, label string) bool {
	for _, s := range w.Scope {
		if s == "*" {
			return true
		}
		k, v, _ := strings.Cut(s, ":")
		if k == kind && (strings.EqualFold(v, key) || strings.EqualFold(v, label)) {
			return true
		}
	}
	return false
}
