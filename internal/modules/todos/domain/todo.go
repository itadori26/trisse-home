package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// MaxTitleLength limits the to-do text.
const MaxTitleLength = 300

// ErrVersionConflict is returned when the to-do was changed since the
// version the client edited (#80).
var ErrVersionConflict = kernel.NewPublicError(kernel.ErrConflict, "The to-do was changed elsewhere in the meantime.")

// ErrTodoNotFound is returned for unknown to-do IDs.
var ErrTodoNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The to-do does not exist.")

// Todo is one task. Due is a calendar date without time (zero = no date).
type Todo struct {
	id          string
	title       string
	due         time.Time
	position    int
	createdAt   time.Time
	completedAt time.Time
	repeat      Recurrence
	reminder    Reminder
	remindedFor time.Time // due date the reminder was sent for
	// version counts the stored changes, starting at 1 (#80).
	version int
}

// errReminderNeedsDue rejects reminders on to-dos without a due date.
var errReminderNeedsDue = kernel.NewValidationError("reminder", "A reminder needs a due date.")

// Schedule sets recurrence and reminder (#68).
func (t *Todo) Schedule(repeat Recurrence, reminder Reminder) error {
	if reminder.Set() && t.due.IsZero() {
		return errReminderNeedsDue
	}
	t.repeat, t.reminder = repeat, reminder
	return nil
}

// WithState restores schedule fields from persisted state.
func (t Todo) WithState(repeat Recurrence, reminder Reminder, remindedFor time.Time) Todo {
	t.repeat, t.reminder, t.remindedFor = repeat, reminder, remindedFor
	return t
}

// Repeat returns the recurrence rule.
func (t Todo) Repeat() Recurrence { return t.repeat }

// Reminder returns the reminder.
func (t Todo) Reminder() Reminder { return t.reminder }

// RemindedFor returns the due date the reminder was last sent for.
func (t Todo) RemindedFor() time.Time { return t.remindedFor }

// ReminderDue reports whether the reminder of an open to-do should be sent
// now; it is sent once per due date.
func (t Todo) ReminderDue(now time.Time, loc *time.Location) bool {
	if t.Done() || !t.reminder.Set() || t.due.IsZero() || t.remindedFor.Equal(t.due) {
		return false
	}
	return !now.Before(t.reminder.FireAt(t.due, loc))
}

// MarkReminded records that the reminder for the current due date was sent.
func (t *Todo) MarkReminded() { t.remindedFor = t.due }

// NextOccurrence returns the follow-up of a completed recurring to-do with
// the same text, rule and reminder; ok is false for one-off to-dos.
// doneOn is the completion day in the user's time zone.
func (t Todo) NextOccurrence(id string, position int, doneOn, now time.Time) (Todo, bool) {
	if !t.repeat.Repeats() {
		return Todo{}, false
	}
	next := Todo{id: id, title: t.title, position: position, createdAt: now, repeat: t.repeat, reminder: t.reminder}
	next.due = t.repeat.Next(t.due, doneOn)
	return next, true
}

// NewTodo validates and creates an open to-do.
func NewTodo(id, title string, due time.Time, position int, now time.Time) (Todo, error) {
	t := Todo{id: id, position: position, createdAt: now, version: 1}
	if err := t.Edit(title, due); err != nil {
		return Todo{}, err
	}
	return t, nil
}

// RestoreTodo rebuilds a to-do from persisted state (version 1; see WithVersion).
func RestoreTodo(id, title string, due time.Time, position int, createdAt, completedAt time.Time) Todo {
	return Todo{id: id, title: title, due: due, position: position, createdAt: createdAt, completedAt: completedAt, version: 1}
}

// WithVersion returns the to-do with its stored version.
func (t Todo) WithVersion(v int) Todo {
	t.version = v
	return t
}

// Version is the number of stored changes (#80).
func (t Todo) Version() int { return t.version }

// CheckVersion fails with ErrVersionConflict when the client edited another
// version than the stored one; 0 means the client did not say (no check).
func (t Todo) CheckVersion(expected int) error {
	if expected != 0 && expected != t.version {
		return ErrVersionConflict
	}
	return nil
}

// Bump counts a change before an existing to-do is stored again; the
// repository only stores it if the stored version is still the one before.
func (t *Todo) Bump() { t.version++ }

// Edit changes text and due date.
func (t *Todo) Edit(title string, due time.Time) error {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLength {
		return kernel.NewValidationError("title", "The to-do must have between 1 and 300 characters.")
	}
	t.title = title
	if !due.IsZero() {
		y, m, d := due.Date()
		due = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	t.due = due
	return nil
}

// SetDone marks the to-do as done (or open again).
func (t *Todo) SetDone(done bool, now time.Time) {
	if done && t.completedAt.IsZero() {
		t.completedAt = now
	}
	if !done {
		t.completedAt = time.Time{}
	}
}

// MoveTo sets the sort position.
func (t *Todo) MoveTo(position int) { t.position = position }

// IsOverdue reports whether an open to-do's due date lies before today.
func (t Todo) IsOverdue(today time.Time) bool {
	if t.Done() || t.due.IsZero() {
		return false
	}
	y, m, d := today.Date()
	return t.due.Before(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

// ID returns the identifier.
func (t Todo) ID() string { return t.id }

// Title returns the text.
func (t Todo) Title() string { return t.title }

// Due returns the due date (zero if none).
func (t Todo) Due() time.Time { return t.due }

// Position returns the sort position.
func (t Todo) Position() int { return t.position }

// CreatedAt returns the creation time.
func (t Todo) CreatedAt() time.Time { return t.createdAt }

// CompletedAt returns when the to-do was completed (zero if open).
func (t Todo) CompletedAt() time.Time { return t.completedAt }

// Done reports whether the to-do is completed.
func (t Todo) Done() bool { return !t.completedAt.IsZero() }
