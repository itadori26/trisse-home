package domain

import (
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// MaxTabNameLength limits tab names so they fit into the tab bar.
const MaxTabNameLength = 40

// Tab is a page of the dashboard, e.g. "Home" or "Server".
type Tab struct {
	id       string
	name     string
	position int
}

// NewTab creates a tab after validating its name.
func NewTab(id, name string, position int) (Tab, error) {
	t := Tab{id: id, position: position}
	if err := t.Rename(name); err != nil {
		return Tab{}, err
	}
	return t, nil
}

// RestoreTab rebuilds a tab from persisted state.
func RestoreTab(id, name string, position int) Tab {
	return Tab{id: id, name: name, position: position}
}

// ID returns the identifier.
func (t Tab) ID() string { return t.id }

// Name returns the display name.
func (t Tab) Name() string { return t.name }

// Position returns the zero-based order in the tab bar.
func (t Tab) Position() int { return t.position }

// Rename changes the name after validation.
func (t *Tab) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return kernel.NewValidationError("name", "The tab name must not be empty.")
	}
	if utf8.RuneCountInString(name) > MaxTabNameLength {
		return kernel.NewValidationError("name", "The tab name must be at most 40 characters long.")
	}
	t.name = name
	return nil
}

// MoveTo sets the position in the tab bar.
func (t *Tab) MoveTo(position int) { t.position = position }
