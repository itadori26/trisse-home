package domain

import (
	"cmp"
	"slices"
	"strings"
	"time"
	"unicode"

	"trisse-home/internal/shared/kernel"
)

// Limits.
const (
	MinQueryLength = 2
	MaxDepartures  = 20
	maxStopIDLen   = 200
)

// ErrInvalidStop is returned for malformed stop IDs.
var ErrInvalidStop = kernel.NewValidationError("stopId", "Please choose a stop in the widget settings.")

// ValidStopID reports whether id can be passed to the provider.
func ValidStopID(id string) bool {
	return id != "" && len(id) <= maxStopIDLen && !strings.ContainsFunc(id, unicode.IsControl)
}

// Stop is a public transport stop.
type Stop struct {
	ID   string
	Name string
	Lat  float64
	Lon  float64
}

// Departure is one departure at a stop.
type Departure struct {
	Line      string
	Headsign  string
	Mode      string
	Planned   time.Time
	Expected  time.Time
	RealTime  bool
	Track     string
	Cancelled bool
}

// DelayMinutes is the real-time delay (0 without real-time data).
func (d Departure) DelayMinutes() int {
	if !d.RealTime || d.Expected.IsZero() {
		return 0
	}
	return int(d.Expected.Sub(d.Planned).Round(time.Minute).Minutes())
}

// When is the expected time, or the planned time without real-time data.
func (d Departure) When() time.Time {
	if d.RealTime && !d.Expected.IsZero() {
		return d.Expected
	}
	return d.Planned
}

// Upcoming drops departures that already left (more than a minute ago,
// cancelled ones are kept until their planned time), sorts by the expected
// time and limits the list to n.
func Upcoming(ds []Departure, now time.Time, n int) []Departure {
	out := make([]Departure, 0, len(ds))
	for _, d := range ds {
		if d.When().After(now.Add(-time.Minute)) {
			out = append(out, d)
		}
	}
	slices.SortStableFunc(out, func(a, b Departure) int { return cmp.Compare(a.When().Unix(), b.When().Unix()) })
	n = max(1, min(n, MaxDepartures))
	if len(out) > n {
		out = out[:n]
	}
	return out
}
