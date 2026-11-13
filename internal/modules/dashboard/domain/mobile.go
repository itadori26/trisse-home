package domain

import (
	"slices"

	"trisse-home/internal/shared/kernel"
)

// MobileEntry places one widget in the single-column phone layout (#92).
type MobileEntry struct {
	WidgetID string
	// Hidden hides the widget on phones only.
	Hidden bool
	// Rows is the minimum height in grid rows (1–MaxWidgetRows).
	Rows int
}

// MobileLayout is the phone layout of a tab: widgets in reading order.
// A tab without a stored layout derives it from the desktop grid.
type MobileLayout struct {
	entries []MobileEntry
	custom  bool
}

// Entries returns the entries in display order.
func (m MobileLayout) Entries() []MobileEntry { return slices.Clone(m.entries) }

// Custom reports whether the layout was edited (false = derived).
func (m MobileLayout) Custom() bool { return m.custom }

// readingOrder sorts widgets top-left to bottom-right of the desktop grid.
func readingOrder(widgets []Widget) []Widget {
	out := slices.Clone(widgets)
	slices.SortStableFunc(out, func(a, b Widget) int {
		if a.rect.Y != b.rect.Y {
			return a.rect.Y - b.rect.Y
		}
		return a.rect.X - b.rect.X
	})
	return out
}

// DeriveMobileLayout builds the phone layout from the desktop grid:
// reading order, desktop height, nothing hidden.
func DeriveMobileLayout(widgets []Widget) MobileLayout {
	var m MobileLayout
	for _, w := range readingOrder(widgets) {
		m.entries = append(m.entries, MobileEntry{WidgetID: w.id, Rows: w.rect.H})
	}
	return m
}

// ResolveMobileLayout applies a stored layout to the current widgets:
// entries of deleted widgets are dropped and widgets added later are
// appended in reading order. A nil stored layout derives a new one.
func ResolveMobileLayout(stored []MobileEntry, widgets []Widget) MobileLayout {
	if stored == nil {
		return DeriveMobileLayout(widgets)
	}
	exists := make(map[string]bool, len(widgets))
	for _, w := range widgets {
		exists[w.id] = true
	}
	m := MobileLayout{custom: true}
	seen := map[string]bool{}
	for _, e := range stored {
		if exists[e.WidgetID] && !seen[e.WidgetID] {
			seen[e.WidgetID] = true
			e.Rows = min(max(e.Rows, 1), MaxWidgetRows)
			m.entries = append(m.entries, e)
		}
	}
	for _, w := range readingOrder(widgets) {
		if !seen[w.id] {
			m.entries = append(m.entries, MobileEntry{WidgetID: w.id, Rows: w.rect.H})
		}
	}
	return m
}

// NewMobileLayout validates an edited layout against the tab's widgets:
// only widgets of the tab, each at most once, heights within the grid.
// Widgets missing from the list are appended by ResolveMobileLayout.
func NewMobileLayout(entries []MobileEntry, widgets []Widget) ([]MobileEntry, error) {
	exists := make(map[string]bool, len(widgets))
	for _, w := range widgets {
		exists[w.id] = true
	}
	seen := map[string]bool{}
	out := make([]MobileEntry, 0, len(entries))
	for _, e := range entries {
		if !exists[e.WidgetID] {
			return nil, ErrWidgetNotFound
		}
		if seen[e.WidgetID] {
			return nil, kernel.NewValidationError("items", "Every widget may appear only once.")
		}
		if e.Rows < 1 || e.Rows > MaxWidgetRows {
			return nil, kernel.NewValidationError("h", "The height must be between 1 and 24 rows.")
		}
		seen[e.WidgetID] = true
		out = append(out, e)
	}
	return out, nil
}
