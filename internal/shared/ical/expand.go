package ical

import (
	"strconv"
	"strings"
	"time"
)

// maxOccurrences bounds the expansion of one rule.
const maxOccurrences = 1000

// Expand returns the occurrences of e that overlap [from, to). Supported:
// FREQ=DAILY|WEEKLY|MONTHLY|YEARLY with INTERVAL, COUNT, UNTIL and EXDATE.
// Rules with BY… parts are not expanded (only the first occurrence is
// returned) rather than guessed.
func Expand(e Event, from, to time.Time) []Event {
	if e.RRule == "" {
		if overlaps(e, from, to) {
			return []Event{e}
		}
		return nil
	}
	rule := map[string]string{}
	for part := range strings.SplitSeq(e.RRule, ";") {
		k, v, _ := strings.Cut(part, "=")
		rule[strings.ToUpper(k)] = strings.ToUpper(v)
	}
	for k := range rule {
		if strings.HasPrefix(k, "BY") && (k != "BYDAY" || !singleWeekday(rule[k], e.Start)) {
			if overlaps(e, from, to) {
				return []Event{e}
			}
			return nil
		}
	}
	interval, _ := strconv.Atoi(rule["INTERVAL"])
	interval = max(interval, 1)
	count, _ := strconv.Atoi(rule["COUNT"])
	var until time.Time
	if u := rule["UNTIL"]; u != "" {
		if t, err := time.Parse("20060102T150405Z", u); err == nil {
			until = t
		} else if t, err := time.Parse("20060102", u); err == nil {
			until = t.Add(24*time.Hour - time.Nanosecond)
		}
	}
	step := func(t time.Time, n int) time.Time {
		switch rule["FREQ"] {
		case "DAILY":
			return t.AddDate(0, 0, n*interval)
		case "WEEKLY":
			return t.AddDate(0, 0, 7*n*interval)
		case "MONTHLY":
			return t.AddDate(0, n*interval, 0)
		case "YEARLY":
			return t.AddDate(n*interval, 0, 0)
		default:
			return time.Time{}
		}
	}
	duration := e.End.Sub(e.Start)
	var out []Event
	for n := 0; n < maxOccurrences; n++ {
		start := step(e.Start, n)
		if start.IsZero() || (count > 0 && n >= count) || (!until.IsZero() && start.After(until)) || !start.Before(to) {
			break
		}
		if excluded(start, e.ExDates) {
			continue
		}
		occ := e
		occ.Start, occ.End, occ.RRule, occ.ExDates = start, start.Add(duration), "", nil
		if overlaps(occ, from, to) {
			out = append(out, occ)
		}
	}
	return out
}

func overlaps(e Event, from, to time.Time) bool { return e.Start.Before(to) && e.End.After(from) }

func excluded(t time.Time, ex []time.Time) bool {
	for _, x := range ex {
		if x.Equal(t) {
			return true
		}
	}
	return false
}

// singleWeekday accepts BYDAY when it only repeats the weekday of start
// (common in exported weekly rules), which needs no special handling.
func singleWeekday(byday string, start time.Time) bool {
	days := []string{"SU", "MO", "TU", "WE", "TH", "FR", "SA"}
	return byday == days[start.Weekday()]
}
