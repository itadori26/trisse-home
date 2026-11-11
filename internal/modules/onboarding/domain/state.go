package domain

import "time"

// State tracks the guided tour. The zero value means "not completed yet",
// which is what a fresh installation and existing accounts start with.
type State struct {
	completedAt time.Time
}

// RestoreState rebuilds a state from persisted data (zero time = pending).
func RestoreState(completedAt time.Time) State { return State{completedAt: completedAt} }

// Completed reports whether the tour has been finished or skipped.
func (s State) Completed() bool { return !s.completedAt.IsZero() }

// CompletedAt returns when the tour was completed (zero if pending).
func (s State) CompletedAt() time.Time { return s.completedAt }

// Complete marks the tour as done. Completing twice keeps the first time.
func (s *State) Complete(now time.Time) {
	if s.completedAt.IsZero() {
		s.completedAt = now
	}
}

// Restart makes the tour show again.
func (s *State) Restart() { s.completedAt = time.Time{} }
