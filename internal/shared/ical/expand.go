package ical

import (
	"slices"
	"strconv"
	"strings"
	"time"
)

// maxOccurrences bounds the expansion of one rule.
const maxOccurrences = 1000

// Expand returns the occurrences of e that overlap [from, to). Supported:
// FREQ=DAILY|WEEKLY|MONTHLY|YEARLY with INTERVAL, COUNT, UNTIL and EXDATE,
// and weekly rules on several weekdays (BYDAY=MO,WE as used by timetables).
// Other BY… parts are not expanded (only the first occurrence is returned)
// rather than guessed.
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
	weekdays, multiDay := weeklyDays(rule)
	for k := range rule {
		if strings.HasPrefix(k, "BY") && (k != "BYDAY" || (!multiDay && !singleWeekday(rule[k], e.Start))) {
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
	emitted := 0
	// emit adds one occurrence; it reports false once the rule is exhausted.
	emit := func(start time.Time) bool {
		if (count > 0 && emitted >= count) || (!until.IsZero() && start.After(until)) || !start.Before(to) {
			return false
		}
		emitted++
		if excluded(start, e.ExDates) {
			return true
		}
		occ := e
		occ.Start, occ.End, occ.RRule, occ.ExDates = start, start.Add(duration), "", nil
		if overlaps(occ, from, to) {
			out = append(out, occ)
		}
		return true
	}
	for n := 0; n < maxOccurrences; n++ {
		start := step(e.Start, n)
		if start.IsZero() {
			break
		}
		if !multiDay {
			if !emit(start) {
				break
			}
			continue
		}
		// Weekly on several days: the days of the week that starts on the
		// Monday of this step, never before DTSTART (RFC 5545 counts DTSTART
		// as the first occurrence).
		monday := start.AddDate(0, 0, -((int(start.Weekday()) + 6) % 7))
		exhausted := false
		for _, wd := range weekdays {
			day := monday.AddDate(0, 0, (int(wd)+6)%7)
			if day.Before(e.Start) {
				continue
			}
			if !emit(day) {
				exhausted = true
				break
			}
		}
		if exhausted {
			break
		}
	}
	return out
}

// weeklyDays returns the weekdays of FREQ=WEEKLY;BYDAY=MO,WE in week order
// (Monday first); ordinals such as 1TU are monthly rules and not supported.
func weeklyDays(rule map[string]string) ([]time.Weekday, bool) {
	byday := rule["BYDAY"]
	if rule["FREQ"] != "WEEKLY" || byday == "" {
		return nil, false
	}
	names := map[string]time.Weekday{"MO": time.Monday, "TU": time.Tuesday, "WE": time.Wednesday, "TH": time.Thursday,
		"FR": time.Friday, "SA": time.Saturday, "SU": time.Sunday}
	var days []time.Weekday
	for part := range strings.SplitSeq(byday, ",") {
		wd, ok := names[part]
		if !ok {
			return nil, false
		}
		days = append(days, wd)
	}
	slices.SortFunc(days, func(a, b time.Weekday) int { return (int(a)+6)%7 - (int(b)+6)%7 })
	return days, len(days) > 1
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
