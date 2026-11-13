package domain

import (
	"cmp"
	"net/url"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// State is a German federal state (ISO 3166-2:DE subdivision code).
type State string

// States in the order of the selection.
var States = []State{"BW", "BY", "BE", "BB", "HB", "HH", "HE", "MV", "NI", "NW", "RP", "SL", "SN", "ST", "SH", "TH"}

// DefaultState is Lower Saxony.
const DefaultState State = "NI"

// ParseState returns the state of code, or DefaultState for an unknown code.
func ParseState(code string) State {
	s := State(strings.ToUpper(strings.TrimSpace(code)))
	if slices.Contains(States, s) {
		return s
	}
	return DefaultState
}

// Kind of an entry.
type Kind string

// Kinds.
const (
	KindPublic Kind = "public"
	KindSchool Kind = "school"
)

// Holiday is a public holiday. Key names it independently of the language
// (e.g. "good_friday"); the UI translates it.
type Holiday struct {
	Date time.Time
	Key  string
}

// SchoolHoliday is a period of school holidays; Start and End are dates
// (midnight UTC), End inclusive.
type SchoolHoliday struct {
	Name       string
	Start, End time.Time
}

// Date returns the calendar date of t as midnight UTC.
func Date(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// Easter returns Easter Sunday of year (Gregorian, anonymous algorithm).
func Easter(year int) time.Time {
	a := year % 19
	b, c := year/100, year%100
	d, e := b/4, b%4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i, k := c/4, c%4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	dayOf := (h+l-7*m+114)%31 + 1
	return day(year, time.Month(month), dayOf)
}

type rule struct {
	key    string
	states []State // nil = all states
	since  int     // first year (0 = always)
	date   func(year int, easter time.Time) time.Time
}

func fixed(m time.Month, d int) func(int, time.Time) time.Time {
	return func(y int, _ time.Time) time.Time { return day(y, m, d) }
}

func easter(offset int) func(int, time.Time) time.Time {
	return func(_ int, e time.Time) time.Time { return e.AddDate(0, 0, offset) }
}

// repentance is Buß- und Bettag: the Wednesday before 23 November.
func repentance(y int, _ time.Time) time.Time {
	d := day(y, time.November, 22)
	for d.Weekday() != time.Wednesday {
		d = d.AddDate(0, 0, -1)
	}
	return d
}

// rules are the statewide public holidays. Holidays that only apply in some
// municipalities (Fronleichnam in parts of Saxony and Thuringia, Mariä
// Himmelfahrt in Bavaria, Augsburger Friedensfest) are left out.
var rules = []rule{
	{key: "new_year", date: fixed(time.January, 1)},
	{key: "epiphany", states: []State{"BW", "BY", "ST"}, date: fixed(time.January, 6)},
	{key: "womens_day", states: []State{"BE"}, since: 2019, date: fixed(time.March, 8)},
	{key: "womens_day", states: []State{"MV"}, since: 2023, date: fixed(time.March, 8)},
	{key: "good_friday", date: easter(-2)},
	{key: "easter_sunday", states: []State{"BB"}, date: easter(0)},
	{key: "easter_monday", date: easter(1)},
	{key: "labour_day", date: fixed(time.May, 1)},
	{key: "ascension", date: easter(39)},
	{key: "whit_sunday", states: []State{"BB"}, date: easter(49)},
	{key: "whit_monday", date: easter(50)},
	{key: "corpus_christi", states: []State{"BW", "BY", "HE", "NW", "RP", "SL"}, date: easter(60)},
	{key: "assumption", states: []State{"SL"}, date: fixed(time.August, 15)},
	{key: "childrens_day", states: []State{"TH"}, since: 2019, date: fixed(time.September, 20)},
	{key: "german_unity", date: fixed(time.October, 3)},
	{key: "reformation", states: []State{"BB", "HB", "HH", "MV", "NI", "SN", "ST", "SH", "TH"}, since: 2018, date: fixed(time.October, 31)},
	{key: "reformation", states: []State{"BB", "MV", "SN", "ST", "TH"}, date: fixed(time.October, 31)},
	{key: "all_saints", states: []State{"BW", "BY", "NW", "RP", "SL"}, date: fixed(time.November, 1)},
	{key: "repentance", states: []State{"SN"}, date: repentance},
	{key: "christmas_day", date: fixed(time.December, 25)},
	{key: "boxing_day", date: fixed(time.December, 26)},
}

// PublicHolidays returns the statewide public holidays of state in year,
// in date order.
func PublicHolidays(year int, state State) []Holiday {
	e := Easter(year)
	var out []Holiday
	for _, r := range rules {
		if (r.states != nil && !slices.Contains(r.states, state)) || year < r.since {
			continue
		}
		d := r.date(year, e)
		if slices.ContainsFunc(out, func(h Holiday) bool { return h.Key == r.key && h.Date.Equal(d) }) {
			continue
		}
		out = append(out, Holiday{Date: d, Key: r.key})
	}
	slices.SortStableFunc(out, func(a, b Holiday) int { return a.Date.Compare(b.Date) })
	return out
}

// Entry is one line of the widget: a public holiday or a school holiday period.
type Entry struct {
	Kind Kind
	// Key is set for public holidays, Name for school holidays.
	Key, Name  string
	Start, End time.Time
	// Current is true while today lies within the entry.
	Current bool
	// DaysUntil is 0 for today and current entries.
	DaysUntil int
}

// MaxEntries bounds the list.
const MaxEntries = 20

// Upcoming merges the public holidays of state (this and next year) and
// the school holidays into the next n entries from today on; current
// school holidays come first.
func Upcoming(today time.Time, state State, school []SchoolHoliday, n int) []Entry {
	today = Date(today)
	n = min(max(n, 1), MaxEntries)
	var out []Entry
	for _, y := range []int{today.Year(), today.Year() + 1} {
		for _, h := range PublicHolidays(y, state) {
			if h.Date.Before(today) {
				continue
			}
			out = append(out, Entry{Kind: KindPublic, Key: h.Key, Start: h.Date, End: h.Date, Current: h.Date.Equal(today), DaysUntil: days(today, h.Date)})
		}
	}
	for _, s := range school {
		if s.End.Before(today) {
			continue
		}
		current := !s.Start.After(today)
		out = append(out, Entry{Kind: KindSchool, Name: s.Name, Start: s.Start, End: s.End, Current: current, DaysUntil: max(0, days(today, s.Start))})
	}
	slices.SortStableFunc(out, func(a, b Entry) int {
		return cmp.Or(cmp.Compare(a.DaysUntil, b.DaysUntil), cmp.Compare(kindOrder(a.Kind), kindOrder(b.Kind)))
	})
	return out[:min(n, len(out))]
}

func kindOrder(k Kind) int {
	if k == KindSchool {
		return 0
	}
	return 1
}

func days(from, to time.Time) int { return int(to.Sub(from).Hours() / 24) }

// Settings of the module: the optional ICS link of the school holidays.
type Settings struct {
	SchoolURL string
}

// NewSettings validates the link; empty disables school holidays.
// webcal:// becomes https://.
func NewSettings(raw string) (Settings, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Settings{}, nil
	}
	if strings.HasPrefix(strings.ToLower(raw), "webcal://") {
		raw = "https://" + raw[len("webcal://"):]
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("schoolUrl", "The ICS link must start with https://, http:// or webcal://.")
	}
	return Settings{SchoolURL: u.String()}, nil
}
