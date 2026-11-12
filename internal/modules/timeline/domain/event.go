package domain

import (
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Retention limits (days).
const (
	DefaultRetentionDays = 30
	MaxRetentionDays     = 365
	maxMessage           = 300
)

// Severity of an event.
type Severity string

// Severities.
const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event is one entry of the timeline.
type Event struct {
	ID       string
	At       time.Time
	Source   string // "service", "container", "backup", "check", "tls", "disk", "diskhealth", "security", …
	Severity Severity
	Message  string
	// Subject identifies the object, e.g. the service ID, for filtering.
	Subject string
}

// NewEvent normalises an event.
func NewEvent(id string, at time.Time, source string, sev Severity, subject, message string) Event {
	message = strings.TrimSpace(message)
	if len(message) > maxMessage {
		message = message[:maxMessage]
	}
	return Event{ID: id, At: at, Source: source, Severity: sev, Subject: subject, Message: message}
}

// criticalKinds are the observation kinds whose problems are critical.
var criticalKinds = map[string]bool{"service": true, "backup": true, "check": true, "diskhealth": true}

// FromObservation describes a confirmed state change.
func FromObservation(id string, at time.Time, o kernel.Observation, healthy bool) Event {
	sev := SeverityWarning
	if criticalKinds[o.Kind] {
		sev = SeverityCritical
	}
	var msg string
	switch {
	case healthy:
		sev, msg = SeverityInfo, o.Label+" is OK again"
	case o.Kind == "service":
		msg = o.Label + " is offline"
	case o.Kind == "disk":
		msg = fmt.Sprintf("Disk %s is %.0f %% full", o.Label, o.Value)
	case o.Kind == "tls":
		msg = fmt.Sprintf("Certificate of %s expires in %.0f days", o.Label, o.Value)
	case o.Kind == "forecast":
		msg = fmt.Sprintf("Disk %s is full in about %.0f days", o.Label, o.Value)
	case o.Kind == "speedtest":
		msg = "Internet is slower than the threshold"
	case o.Kind == "diskhealth":
		msg = "Disk " + o.Label + " is degrading"
	default:
		msg = o.Label + " needs attention"
	}
	if !healthy && o.Detail != "" {
		msg += ": " + o.Detail
	}
	return NewEvent(id, at, o.Kind, sev, o.Key, msg)
}

// HealthyFor judges threshold observations like the notification defaults.
func HealthyFor(o kernel.Observation) bool {
	switch o.Kind {
	case "disk":
		return o.Value < 90
	case "tls":
		return o.Value > 14
	}
	return o.Healthy
}

// ValidRetention checks the configured retention.
func ValidRetention(days int) error {
	if days < 1 || days > MaxRetentionDays {
		return kernel.NewValidationError("retentionDays", "The retention must be between 1 and 365 days.")
	}
	return nil
}

// Track debounces observations (two confirmations, baseline on first).
type Track struct {
	Known     bool
	Healthy   bool
	candidate bool
	count     int
}

// Advance returns the new track and whether the state change is confirmed.
func (t Track) Advance(healthy bool) (Track, bool) {
	if !t.Known {
		return Track{Known: true, Healthy: healthy}, false
	}
	if healthy == t.Healthy {
		t.count = 0
		return t, false
	}
	if t.count == 0 || t.candidate != healthy {
		t.candidate, t.count = healthy, 0
	}
	t.count++
	if t.count < 2 {
		return t, false
	}
	return Track{Known: true, Healthy: healthy}, true
}
