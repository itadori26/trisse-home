package domain

import (
	"slices"
	"time"
)

// SectionKind names a part of the briefing (#52). Each section has one
// source module; it is empty when that module is disabled.
type SectionKind string

// Sections in their default order: problems first, then the day's plans.
const (
	SectionWarnings  SectionKind = "warnings"  // weather warnings
	SectionIncidents SectionKind = "incidents" // overnight failures from the timeline
	SectionServices  SectionKind = "services"  // services that are down now
	SectionTodos     SectionKind = "todos"     // due and overdue to-dos
	SectionEvents    SectionKind = "events"    // next appointments
	SectionWaste     SectionKind = "waste"     // collections today and tomorrow
)

// DefaultOrder is the order of a new briefing.
var DefaultOrder = []SectionKind{SectionWarnings, SectionIncidents, SectionServices, SectionTodos, SectionEvents, SectionWaste}

// MaxItemsPerSection keeps the briefing compact; the rest is counted.
const MaxItemsPerSection = 5

// Severity of an item: warnings and failures stand out.
type Severity int

// Severities from least to most urgent.
const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityCritical
)

// Item is one line of a section.
type Item struct {
	Text string
	// Detail is secondary text, e.g. a location or "overdue".
	Detail   string
	At       time.Time
	Severity Severity
}

// Section is a non-empty part of the briefing; More counts the items
// that did not fit.
type Section struct {
	Kind  SectionKind
	Items []Item
	More  int
}

// Briefing is the summary of the day: the enabled, non-empty sections in
// the configured order.
type Briefing struct {
	Sections []Section
}

// Empty reports whether there is nothing to report.
func (b Briefing) Empty() bool { return len(b.Sections) == 0 }

// LayoutEntry switches one section on or off.
type LayoutEntry struct {
	Kind    SectionKind
	Enabled bool
}

// Layout is the order and visibility of the sections.
type Layout []LayoutEntry

// DefaultLayout shows every section in the default order.
func DefaultLayout() Layout {
	out := make(Layout, 0, len(DefaultOrder))
	for _, k := range DefaultOrder {
		out = append(out, LayoutEntry{Kind: k, Enabled: true})
	}
	return out
}

// NormaliseLayout drops unknown and duplicate kinds and appends missing
// ones (enabled), so a stored layout survives new section kinds.
func NormaliseLayout(in []LayoutEntry) Layout {
	out := Layout{}
	seen := map[SectionKind]bool{}
	for _, e := range in {
		if slices.Contains(DefaultOrder, e.Kind) && !seen[e.Kind] {
			seen[e.Kind] = true
			out = append(out, e)
		}
	}
	for _, k := range DefaultOrder {
		if !seen[k] {
			out = append(out, LayoutEntry{Kind: k, Enabled: true})
		}
	}
	return out
}

// Compose builds the briefing from the items of each section: enabled
// sections in layout order, empty ones left out. Within a section the
// most urgent items come first, then the earliest.
func Compose(layout Layout, items map[SectionKind][]Item) Briefing {
	var b Briefing
	for _, e := range NormaliseLayout(layout) {
		list := slices.Clone(items[e.Kind])
		if !e.Enabled || len(list) == 0 {
			continue
		}
		slices.SortStableFunc(list, func(a, b Item) int {
			if a.Severity != b.Severity {
				return int(b.Severity - a.Severity)
			}
			return a.At.Compare(b.At)
		})
		s := Section{Kind: e.Kind, Items: list}
		if len(list) > MaxItemsPerSection {
			s.Items, s.More = list[:MaxItemsPerSection], len(list)-MaxItemsPerSection
		}
		b.Sections = append(b.Sections, s)
	}
	return b
}

// IncidentsSince is the start of "since yesterday evening": 18:00 of the
// previous day in loc.
func IncidentsSince(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d-1, 18, 0, 0, 0, now.Location())
}

// EndOfTomorrow is the end of the window for appointments and collections.
func EndOfTomorrow(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d+2, 0, 0, 0, 0, now.Location())
}
