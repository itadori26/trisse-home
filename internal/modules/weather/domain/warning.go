package domain

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

// Severity of an official warning (CAP).
type Severity string

// Severities, ordered by rank.
const (
	SeverityMinor    Severity = "minor"
	SeverityModerate Severity = "moderate"
	SeveritySevere   Severity = "severe"
	SeverityExtreme  Severity = "extreme"
)

// ParseSeverity maps CAP severities ("Moderate") to Severity.
func ParseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "extreme":
		return SeverityExtreme
	case "severe":
		return SeveritySevere
	case "moderate":
		return SeverityModerate
	default:
		return SeverityMinor
	}
}

func (s Severity) rank() int {
	return slices.Index([]Severity{SeverityMinor, SeverityModerate, SeveritySevere, SeverityExtreme}, s)
}

// Warning is an official weather warning for a location.
type Warning struct {
	ID          string
	Event       string
	Headline    string
	Description string
	Instruction string
	Severity    Severity
	Onset       time.Time
	Expires     time.Time
}

// RelevantWarnings drops expired warnings and duplicates and orders the rest
// by severity (highest first), then by onset.
func RelevantWarnings(ws []Warning, now time.Time) []Warning {
	seen := map[string]bool{}
	out := make([]Warning, 0, len(ws))
	for _, w := range ws {
		if (!w.Expires.IsZero() && !w.Expires.After(now)) || seen[w.ID] {
			continue
		}
		seen[w.ID] = true
		out = append(out, w)
	}
	slices.SortStableFunc(out, func(a, b Warning) int {
		return cmp.Or(cmp.Compare(b.Severity.rank(), a.Severity.rank()), a.Onset.Compare(b.Onset))
	})
	return out
}

// InGermany reports whether the location is covered by the DWD services.
func (l Location) InGermany() bool {
	return l.lat >= 47.2 && l.lat <= 55.1 && l.lon >= 5.8 && l.lon <= 15.1
}

// BBox is a geographic rectangle (WGS84).
type BBox struct {
	MinLat, MinLon, MaxLat, MaxLon float64
}

// RadarBBox returns the area shown by the radar widget, centred on l
// (about 220 × 220 km at German latitudes).
func (l Location) RadarBBox() BBox {
	const dLat, dLon = 1.0, 1.6
	return BBox{MinLat: l.lat - dLat, MinLon: l.lon - dLon, MaxLat: l.lat + dLat, MaxLon: l.lon + dLon}
}
