package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// MaxTitleLength limits the to-do text.
const MaxTitleLength = 300

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
}

// NewTodo validates and creates an open to-do.
func NewTodo(id, title string, due time.Time, position int, now time.Time) (Todo, error) {
	t := Todo{id: id, position: position, createdAt: now}
	if err := t.Edit(title, due); err != nil {
		return Todo{}, err
	}
	return t, nil
}

// RestoreTodo rebuilds a to-do from persisted state.
func RestoreTodo(id, title string, due time.Time, position int, createdAt, completedAt time.Time) Todo {
	return Todo{id: id, title: title, due: due, position: position, createdAt: createdAt, completedAt: completedAt}
}

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
