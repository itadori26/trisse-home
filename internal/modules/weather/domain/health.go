package domain

import (
	"slices"
	"strings"
	"time"
)

// Pollen and UV index of the DWD health forecasts (#56).

// PollenType is one of the eight allergens the DWD forecasts.
type PollenType string

// Pollen types in the order of the season.
const (
	Hazel    PollenType = "hazel"
	Alder    PollenType = "alder"
	Ash      PollenType = "ash"
	Birch    PollenType = "birch"
	Grasses  PollenType = "grasses"
	Rye      PollenType = "rye"
	Mugwort  PollenType = "mugwort"
	Ambrosia PollenType = "ambrosia"
)

// PollenTypes in season order.
var PollenTypes = []PollenType{Hazel, Alder, Ash, Birch, Grasses, Rye, Mugwort, Ambrosia}

// dwdPollenNames maps the German names of the feed.
var dwdPollenNames = map[string]PollenType{
	"hasel": Hazel, "erle": Alder, "esche": Ash, "birke": Birch,
	"graeser": Grasses, "gräser": Grasses, "roggen": Rye, "beifuss": Mugwort, "beifuß": Mugwort, "ambrosia": Ambrosia,
}

// ParsePollenType maps a feed name ("Graeser") or a type ("grasses").
func ParsePollenType(s string) (PollenType, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if t, ok := dwdPollenNames[s]; ok {
		return t, true
	}
	if slices.Contains(PollenTypes, PollenType(s)) {
		return PollenType(s), true
	}
	return "", false
}

// PollenLevel is the DWD index in half steps: 0 = none, 1 = none to low,
// 2 = low, … 6 = high; -1 = no forecast.
type PollenLevel int

// NoPollenData is the level of a missing forecast.
const NoPollenData PollenLevel = -1

var pollenSteps = []string{"0", "0-1", "1", "1-2", "2", "2-3", "3"}

// ParsePollenLevel reads "0", "0-1", … "3"; anything else is NoPollenData.
func ParsePollenLevel(s string) PollenLevel {
	if i := slices.Index(pollenSteps, strings.TrimSpace(s)); i >= 0 {
		return PollenLevel(i)
	}
	return NoPollenData
}

// String is the DWD notation ("1-2"), or "" without data.
func (l PollenLevel) String() string {
	if l < 0 || int(l) >= len(pollenSteps) {
		return ""
	}
	return pollenSteps[l]
}

// Burden classifies the level for colours: none, low, moderate, high.
func (l PollenLevel) Burden() string {
	switch {
	case l < 0:
		return "unknown"
	case l == 0:
		return "none"
	case l <= 2:
		return "low"
	case l <= 4:
		return "moderate"
	default:
		return "high"
	}
}

// PollenDay is the forecast of one allergen.
type PollenDay struct {
	Type            PollenType
	Today, Tomorrow PollenLevel
}

// PollenRegion is a forecast region of the DWD: a part region where the
// DWD splits a region, else the whole region.
type PollenRegion struct {
	ID   int
	Name string
}

// PollenForecast is the forecast of one region.
type PollenForecast struct {
	Region    PollenRegion
	Days      []PollenDay
	UpdatedAt time.Time
}

// Of returns the forecast of the given types (all when types is empty), in season order.
func (f PollenForecast) Of(types []PollenType) []PollenDay {
	out := make([]PollenDay, 0, len(f.Days))
	for _, d := range f.Days {
		if len(types) == 0 || slices.Contains(types, d.Type) {
			out = append(out, d)
		}
	}
	slices.SortStableFunc(out, func(a, b PollenDay) int {
		return slices.Index(PollenTypes, a.Type) - slices.Index(PollenTypes, b.Type)
	})
	return out
}

// UVLevel classifies a UV index (WHO): low 0–2, moderate 3–5, high 6–7,
// very high 8–10, extreme 11+.
func UVLevel(index int) string {
	switch {
	case index < 0:
		return "unknown"
	case index <= 2:
		return "low"
	case index <= 5:
		return "moderate"
	case index <= 7:
		return "high"
	case index <= 10:
		return "very_high"
	default:
		return "extreme"
	}
}

// UVForecast is the UV index forecast of one DWD station city.
type UVForecast struct {
	City            string
	Today, Tomorrow int
	UpdatedAt       time.Time
}
