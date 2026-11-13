// Package ical parses and writes the subset of iCalendar (RFC 5545) that
// trisse-home needs: VEVENTs with dates, times, durations, simple recurrence
// rules and text escaping. It is shared by the adapter rings of the calendar,
// waste and shopping modules.
package ical

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Event is one VEVENT. All-day events have AllDay set and Start/End at
// midnight UTC of their dates; End is exclusive.
type Event struct {
	UID      string
	Summary  string
	Location string
	// Description is the plain text of DESCRIPTION, at most maxDescription runes (#163).
	Description string
	// URL is the raw URL property; callers decide which schemes they show.
	URL    string
	Start  time.Time
	End    time.Time
	AllDay bool
	// RRule is the raw recurrence rule (empty for single events).
	RRule string
	// ExDates are excluded occurrences.
	ExDates []time.Time
}

// icalProperty is one content line, e.g. DTSTART;TZID=Europe/Berlin:20261110T090000.
type icalProperty struct {
	name   string
	params map[string]string
	value  string
}

func parseLine(line string) (icalProperty, bool) {
	head, value, ok := cutUnquoted(line, ':')
	if !ok {
		return icalProperty{}, false
	}
	parts := strings.Split(head, ";")
	p := icalProperty{name: strings.ToUpper(parts[0]), params: map[string]string{}, value: value}
	for _, param := range parts[1:] {
		k, v, _ := strings.Cut(param, "=")
		p.params[strings.ToUpper(k)] = strings.Trim(v, `"`)
	}
	return p, true
}

// cutUnquoted splits at the first sep outside double quotes (parameter values
// such as TZID may be quoted and contain colons).
func cutUnquoted(s string, sep byte) (string, string, bool) {
	quoted := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			quoted = !quoted
		case sep:
			if !quoted {
				return s[:i], s[i+1:], true
			}
		}
	}
	return s, "", false
}

// unfold joins folded lines (RFC 5545 §3.1).
func unfold(data string) []string {
	data = strings.ReplaceAll(data, "\r\n", "\n")
	data = strings.ReplaceAll(data, "\n ", "")
	data = strings.ReplaceAll(data, "\n\t", "")
	return strings.Split(data, "\n")
}

var textUnescaper = strings.NewReplacer(`\n`, "\n", `\N`, "\n", `\,`, ",", `\;`, ";", `\\`, `\`)

// Parse extracts the VEVENTs of an iCalendar object. Floating times
// (without zone) are interpreted in loc. Recurrence rules are not expanded;
// use Expand for that.
func Parse(data string, loc *time.Location) ([]Event, error) {
	var events []Event
	var current map[string]icalProperty
	var exdates []icalProperty
	depth := 0 // nesting inside VEVENT (VALARM components)
	for _, line := range unfold(data) {
		switch {
		case line == "BEGIN:VEVENT":
			current, exdates, depth = map[string]icalProperty{}, nil, 0
		case strings.HasPrefix(line, "BEGIN:") && current != nil:
			depth++
		case strings.HasPrefix(line, "END:") && current != nil && depth > 0:
			depth--
		case line == "END:VEVENT" && current != nil:
			if e, ok := buildEvent(current, exdates, loc); ok {
				events = append(events, e)
			}
			current = nil
		case current != nil && depth == 0:
			if p, ok := parseLine(line); ok {
				if p.name == "EXDATE" {
					exdates = append(exdates, p)
				} else if _, seen := current[p.name]; !seen {
					current[p.name] = p
				}
			}
		}
	}
	if events == nil && !strings.Contains(data, "BEGIN:VCALENDAR") {
		return nil, errors.New("not an iCalendar object")
	}
	return events, nil
}

func buildEvent(props map[string]icalProperty, exdates []icalProperty, loc *time.Location) (Event, bool) {
	if strings.EqualFold(props["STATUS"].value, "CANCELLED") {
		return Event{}, false
	}
	startProp, ok := props["DTSTART"]
	if !ok {
		return Event{}, false
	}
	start, allDay, err := parseDateTime(startProp, loc)
	if err != nil {
		return Event{}, false
	}
	end := start
	switch {
	case props["DTEND"].value != "":
		if e, _, err := parseDateTime(props["DTEND"], loc); err == nil {
			end = e
		}
	case props["DURATION"].value != "":
		if d, err := parseDuration(props["DURATION"].value); err == nil {
			end = start.Add(d)
		}
	case allDay:
		end = start.AddDate(0, 0, 1)
	}
	summary := strings.TrimSpace(textUnescaper.Replace(props["SUMMARY"].value))
	if summary == "" {
		summary = "(no title)"
	}
	e := Event{
		UID: props["UID"].value, Summary: summary, Location: strings.TrimSpace(textUnescaper.Replace(props["LOCATION"].value)),
		Description: truncateRunes(strings.TrimSpace(textUnescaper.Replace(props["DESCRIPTION"].value)), maxDescription),
		URL:         strings.TrimSpace(props["URL"].value),
		Start:       start, End: end, AllDay: allDay, RRule: strings.TrimSpace(props["RRULE"].value),
	}
	for _, ex := range exdates {
		for _, v := range strings.Split(ex.value, ",") {
			if t, _, err := parseDateTime(icalProperty{params: ex.params, value: v}, loc); err == nil {
				e.ExDates = append(e.ExDates, t)
			}
		}
	}
	return e, true
}

func parseDateTime(p icalProperty, loc *time.Location) (time.Time, bool, error) {
	v := strings.TrimSpace(p.value)
	if p.params["VALUE"] == "DATE" || len(v) == len("20060102") {
		t, err := time.Parse("20060102", v)
		return t, true, err
	}
	if strings.HasSuffix(v, "Z") {
		t, err := time.Parse("20060102T150405Z", v)
		return t, false, err
	}
	zone := loc
	if tzid := p.params["TZID"]; tzid != "" {
		if l, err := time.LoadLocation(tzid); err == nil {
			zone = l
		}
	}
	t, err := time.ParseInLocation("20060102T150405", v, zone)
	return t, false, err
}

var durationPattern = regexp.MustCompile(`^([+-])?P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

// parseDuration parses RFC 5545 durations such as PT1H30M or P1D.
func parseDuration(v string) (time.Duration, error) {
	m := durationPattern.FindStringSubmatch(strings.TrimSpace(v))
	if m == nil {
		return 0, errors.New("invalid duration")
	}
	units := []time.Duration{7 * 24 * time.Hour, 24 * time.Hour, time.Hour, time.Minute, time.Second}
	var d time.Duration
	for i, u := range units {
		if m[i+2] == "" {
			continue
		}
		n, _ := strconv.Atoi(m[i+2]) // the pattern guarantees digits
		d += time.Duration(n) * u
	}
	if m[1] == "-" {
		d = -d
	}
	return d, nil
}

// maxDescription bounds the description kept per event; feeds sometimes carry
// whole documents in it.
const maxDescription = 2000

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
