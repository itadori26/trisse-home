package domain

import (
	"strings"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// MaxTabNameLength limits tab names so they fit into the tab bar.
const MaxTabNameLength = 40

// Tab is a page of the dashboard, e.g. "Home" or "Server".
type Tab struct {
	id         string
	name       string
	position   int
	background string
	rules      []TabRule
	mobile     []MobileEntry
	// zoned lays the tab out in zones instead of the free grid (#619).
	zoned bool
	// sample marks an example tab created at setup (#670) that the template
	// choice on first start may replace.
	sample bool
	// access is the owner and visibility (ADR 0018). A tab an admin shares
	// with a group is a team tab (#295): the group sees it read-only.
	access kernel.Access
}

// WithAccess returns the tab with owner and visibility.
func (t Tab) WithAccess(a kernel.Access) Tab {
	t.access = a
	return t
}

// Access returns owner and visibility; empty for a tab not stored yet.
func (t Tab) Access() kernel.Access { return t.access }

// Mobile returns the stored phone layout (nil = derived from the desktop grid, #92).
func (t Tab) Mobile() []MobileEntry { return t.mobile }

// WithMobile returns the tab with a phone layout (validated by
// NewMobileLayout or restored); nil resets it to the derived layout.
func (t Tab) WithMobile(m []MobileEntry) Tab {
	t.mobile = m
	return t
}

// Zoned reports whether the tab is laid out in zones (#619).
func (t Tab) Zoned() bool { return t.zoned }

// WithZones returns the tab with zones switched on or off. The widgets keep
// their grid places, so switching back restores the free grid unchanged.
func (t Tab) WithZones(on bool) Tab {
	t.zoned = on
	return t
}

// Sample reports whether the tab is an example tab from the setup (#670).
func (t Tab) Sample() bool { return t.sample }

// WithSample returns the tab marked as an example tab or not.
func (t Tab) WithSample(on bool) Tab {
	t.sample = on
	return t
}

// Rules returns the default-tab rules (#71).
func (t Tab) Rules() []TabRule { return t.rules }

// WithRules returns the tab with rules (validated by NewTabRules or restored).
func (t Tab) WithRules(r []TabRule) Tab {
	t.rules = r
	return t
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
func RestoreTab(id, name string, position int, background string) Tab {
	return Tab{id: id, name: name, position: position, background: background}
}

// Background returns the tab's background image URL ("" = global one).
func (t Tab) Background() string { return t.background }

// SetBackground sets the tab's background (see kernel.ParseBackground);
// empty uses the global background.
func (t *Tab) SetBackground(raw string) error {
	bg, err := kernel.ParseBackground("background", raw)
	if err != nil {
		return err
	}
	t.background = bg
	return nil
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
