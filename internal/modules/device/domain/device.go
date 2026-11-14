package domain

import (
	"strings"

	"trisse-home/internal/shared/kernel"
)

// Mode is how the appliance is used.
type Mode string

// Modes; empty means not chosen yet (the setup step asks).
const (
	ModeUnset    Mode = ""
	ModeHeadless Mode = "headless"
	ModeKiosk    Mode = "kiosk"
)

// ParseMode validates a requested mode.
func ParseMode(raw string) (Mode, error) {
	switch m := Mode(strings.TrimSpace(raw)); m {
	case ModeHeadless, ModeKiosk:
		return m, nil
	}
	return "", kernel.NewValidationError("mode", "Choose headless or kiosk.")
}

// Phase is where a mode change stands; the root script reports it.
type Phase string

// Phases of a change.
const (
	PhaseIdle       Phase = "idle"       // nothing requested yet
	PhasePending    Phase = "pending"    // requested, the system has not picked it up
	PhaseInstalling Phase = "installing" // the system installs or removes the kiosk
	PhaseActive     Phase = "active"     // the requested mode runs
	PhaseFailed     Phase = "failed"     // see the error
)

// Display is whether a screen is connected (sysfs).
type Display string

// Display states.
const (
	DisplayConnected    Display = "connected"
	DisplayDisconnected Display = "disconnected"
	DisplayUnknown      Display = "unknown"
)

// Progress is what the root script last reported.
type Progress struct {
	Mode  Mode
	Phase Phase
	Error string
}

// State is the display state of the device.
type State struct {
	Requested Mode
	Progress  Progress
	Display   Display
}

// Phase derives the phase: a request the system has not answered is pending.
func (s State) Phase() Phase {
	switch {
	case s.Requested == ModeUnset:
		return PhaseIdle
	case s.Progress.Mode != s.Requested || s.Progress.Phase == "":
		return PhasePending
	}
	return s.Progress.Phase
}

// Errors.
var (
	ErrNotAppliance = kernel.NewPublicError(kernel.ErrNotFound, "Display modes are only available on the trisse-home appliance image.")
	ErrBusy         = kernel.NewPublicError(kernel.ErrConflict, "A change of the display mode is still in progress.")
)

// CanChange reports whether a new mode may be requested now.
func (s State) CanChange() error {
	if p := s.Phase(); p == PhaseInstalling {
		return ErrBusy
	}
	return nil
}
