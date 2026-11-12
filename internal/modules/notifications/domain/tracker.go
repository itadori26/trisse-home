package domain

import (
	"fmt"
	"math"

	"trisse-home/internal/shared/kernel"
)

// Confirmations is how many consecutive observations a new state needs
// before it is notified; it suppresses flapping.
const Confirmations = 2

// Track is the debounced state of one observed key.
type Track struct {
	// Known is false until the first observation established a baseline.
	Known     bool
	Healthy   bool
	candidate bool
	count     int
}

// Advance feeds one observation into the track. It returns true when the
// confirmed state changed and a notification should be sent. The first
// observation only sets the baseline, so restarts do not cause messages.
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
	if t.count < Confirmations {
		return t, false
	}
	return Track{Known: true, Healthy: healthy}, true
}

// Message is a push notification.
type Message struct {
	Title    string
	Body     string
	Priority int
	Tags     []string
	// Type is the alert type ("service", "security", …, "digest").
	Type     string
	Severity Severity
}

// Priorities as defined by ntfy.
const (
	PriorityDefault = 3
	PriorityHigh    = 4
)

// MessageFor describes a confirmed state change of an observation.
func MessageFor(o kernel.Observation, healthy bool) Message {
	if healthy {
		return Message{Title: fmt.Sprintf("%s is OK again", o.Label), Body: recoveredText(o), Priority: PriorityDefault, Tags: []string{"white_check_mark"}}
	}
	return Message{Title: problemTitle(o), Body: o.Detail, Priority: PriorityHigh, Tags: []string{"warning"}}
}

func problemTitle(o kernel.Observation) string {
	switch o.Kind {
	case "service":
		return o.Label + " is offline"
	case "container":
		return "Container " + o.Label + " needs attention"
	case "disk":
		return fmt.Sprintf("Disk %s is %.0f %% full", o.Label, o.Value)
	case "tls":
		return fmt.Sprintf("Certificate of %s expires in %d days", o.Label, int(math.Floor(o.Value)))
	case "backup":
		return "Backup " + o.Label + " needs attention"
	case "diskhealth":
		return "Disk " + o.Label + " is degrading"
	case "forecast":
		return fmt.Sprintf("Disk %s is full in about %.0f days", o.Label, o.Value)
	default:
		return o.Label + " needs attention"
	}
}

func recoveredText(o kernel.Observation) string {
	if o.Detail != "" {
		return o.Detail
	}
	return "The problem is resolved."
}
