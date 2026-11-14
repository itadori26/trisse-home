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
	// version counts the stored changes, starting at 1 (#80).
	version int
}

// ErrVersionConflict is returned when the note was changed since the
// version the client edited (#80).
var ErrVersionConflict = kernel.NewPublicError(kernel.ErrConflict, "The note was changed elsewhere in the meantime.")

// NewNote validates and creates a note.
func NewNote(id, title, body string, pinned bool, now time.Time) (Note, error) {
	n := Note{id: id, createdAt: now, version: 1}
	if err := n.Edit(title, body, pinned, now); err != nil {
		return Note{}, err
	}
	return n, nil
}

// RestoreNote rebuilds a note from persisted state (version 1; see WithVersion).
func RestoreNote(id, title, body string, pinned bool, createdAt, updatedAt time.Time) Note {
	return Note{id: id, title: title, body: body, pinned: pinned, createdAt: createdAt, updatedAt: updatedAt, version: 1}
}

// WithVersion returns the note with its stored version.
func (n Note) WithVersion(v int) Note {
	n.version = v
	return n
}

// Version is the number of stored changes (#80).
func (n Note) Version() int { return n.version }

// CheckVersion fails with ErrVersionConflict when the client edited another
// version than the stored one; 0 means no check.
func (n Note) CheckVersion(expected int) error {
	if expected != 0 && expected != n.version {
		return ErrVersionConflict
	}
	return nil
}

// Bump counts a change before an existing note is stored again.
func (n *Note) Bump() { n.version++ }

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
