package domain

import (
	"strings"
	"unicode"

	"trisse-home/internal/shared/kernel"
)

// CapturePrefixes are the short prefixes that turn a search into quick
// capture (#105): "t Milch kaufen" creates a to-do. They are followed by a
// space, so they never collide with DuckDuckGo !bangs.
type CapturePrefixes struct {
	Todo, Note, Bookmark string
}

// BarOptions configure the search field of the dashboard (#221): whether web
// results open in a new tab and the quick capture prefixes. They used to be
// options of the search widget, which the hero search field replaced.
type BarOptions struct {
	openInNewTab bool
	prefixes     CapturePrefixes
}

// DefaultCapturePrefixes are the prefixes used until the owner changes them.
var DefaultCapturePrefixes = CapturePrefixes{Todo: "t", Note: "n", Bookmark: "b"}

// DefaultBarOptions open web results in the same tab and use the default prefixes.
func DefaultBarOptions() BarOptions { return BarOptions{prefixes: DefaultCapturePrefixes} }

var (
	errPrefixForm = kernel.NewValidationError("capturePrefixes", "A prefix is 1 to 4 letters or digits.")
	errPrefixDup  = kernel.NewValidationError("capturePrefixes", "Each prefix can be used only once.")
)

// NewBarOptions validates the options. Prefixes are trimmed and lower-cased;
// each is 1 to 4 letters or digits, and no two are the same.
func NewBarOptions(openInNewTab bool, p CapturePrefixes) (BarOptions, error) {
	norm := func(s string) (string, error) {
		s = strings.ToLower(strings.TrimSpace(s))
		n := 0
		for _, r := range s {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return "", errPrefixForm
			}
			n++
		}
		if n < 1 || n > 4 {
			return "", errPrefixForm
		}
		return s, nil
	}
	var err error
	out := CapturePrefixes{}
	if out.Todo, err = norm(p.Todo); err != nil {
		return BarOptions{}, err
	}
	if out.Note, err = norm(p.Note); err != nil {
		return BarOptions{}, err
	}
	if out.Bookmark, err = norm(p.Bookmark); err != nil {
		return BarOptions{}, err
	}
	if out.Todo == out.Note || out.Todo == out.Bookmark || out.Note == out.Bookmark {
		return BarOptions{}, errPrefixDup
	}
	return BarOptions{openInNewTab: openInNewTab, prefixes: out}, nil
}

// OpenInNewTab reports whether web results open in a new tab.
func (o BarOptions) OpenInNewTab() bool { return o.openInNewTab }

// Prefixes returns the quick capture prefixes.
func (o BarOptions) Prefixes() CapturePrefixes { return o.prefixes }
