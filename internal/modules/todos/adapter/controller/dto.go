package controller

// TodoInputDTO is the body of POST /api/todos and PUT /api/todos/{id}.
// Due is a date (YYYY-MM-DD) or empty.
type TodoInputDTO struct {
	Title string `json:"title"`
	Due   string `json:"due"`
	Done  bool   `json:"done"`
	// Repeat and Reminder are optional; when omitted the stored values are
	// kept (#68).
	Repeat   *RepeatDTO   `json:"repeat,omitempty"`
	Reminder *ReminderDTO `json:"reminder,omitempty"`
	// ID (create) is a client-generated ID; replaying the create returns the
	// existing to-do (#80).
	ID string `json:"id,omitempty"`
	// Version (update) is the version the client edited; a different stored
	// version answers 409 Conflict. 0/omitted = no check.
	Version int `json:"version,omitempty"`
}

// RepeatDTO is a recurrence: kind "" (none), "daily", "weekly", "monthly"
// or "after" (days after completion); days are weekdays (0 = Sunday) of
// weekly rules.
type RepeatDTO struct {
	Kind  string `json:"kind"`
	Every int    `json:"every"`
	Days  []int  `json:"days"`
}

// ReminderDTO is a reminder at a time ("HH:MM", "" = none) on the due date
// or some days before.
type ReminderDTO struct {
	At         string `json:"at"`
	DaysBefore int    `json:"daysBefore"`
}

// ClearedDTO is returned by POST /api/todos/clear-completed.
type ClearedDTO struct {
	Removed int `json:"removed"`
}
