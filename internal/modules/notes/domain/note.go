// Package domain holds the Note entity.
//
// Ring: Entities.
package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Field limits.
const (
	MaxTitleLength = 200
	MaxBodyBytes   = 100 * 1024
)

// ErrNoteNotFound is returned for unknown note IDs.
var ErrNoteNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The note does not exist.")

// Note is a Markdown note.
type Note struct {
	id        string
	title     string
	body      string
	pinned    bool
	createdAt time.Time
	updatedAt time.Time
}

// NewNote validates and creates a note.
func NewNote(id, title, body string, pinned bool, now time.Time) (Note, error) {
	n := Note{id: id, createdAt: now}
	if err := n.Edit(title, body, pinned, now); err != nil {
		return Note{}, err
	}
	return n, nil
}

// RestoreNote rebuilds a note from persisted state.
func RestoreNote(id, title, body string, pinned bool, createdAt, updatedAt time.Time) Note {
	return Note{id: id, title: title, body: body, pinned: pinned, createdAt: createdAt, updatedAt: updatedAt}
}

// Edit replaces title, body and pin state. A note needs a title or a body;
// an empty title is derived from the first line of the body.
func (n *Note) Edit(title, body string, pinned bool, now time.Time) error {
	title = strings.TrimSpace(title)
	if len(body) > MaxBodyBytes {
		return kernel.NewValidationError("body", "The note must be at most 100 KiB long.")
	}
	if title == "" {
		title = firstLine(body)
	}
	if title == "" {
		return kernel.NewValidationError("title", "A note needs a title or some text.")
	}
	if utf8.RuneCountInString(title) > MaxTitleLength {
		title = string([]rune(title)[:MaxTitleLength])
	}
	n.title, n.body, n.pinned, n.updatedAt = title, body, pinned, now
	return nil
}

func firstLine(body string) string {
	for line := range strings.SplitSeq(body, "\n") {
		if t := strings.TrimSpace(strings.TrimLeft(line, "# ")); t != "" {
			return t
		}
	}
	return ""
}

// ID returns the identifier.
func (n Note) ID() string { return n.id }

// Title returns the title.
func (n Note) Title() string { return n.title }

// Body returns the Markdown text.
func (n Note) Body() string { return n.body }

// Pinned reports whether the note is shown on the dashboard widget.
func (n Note) Pinned() bool { return n.pinned }

// CreatedAt returns the creation time.
func (n Note) CreatedAt() time.Time { return n.createdAt }

// UpdatedAt returns the time of the last edit.
func (n Note) UpdatedAt() time.Time { return n.updatedAt }
