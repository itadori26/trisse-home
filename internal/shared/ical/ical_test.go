package ical

import (
	"strings"
	"testing"
	"time"
)

const wasteICS = "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:a\r\nSUMMARY:Biotonne\r\nDTSTART;VALUE=DATE:20261102\r\n" +
	"RRULE:FREQ=WEEKLY;INTERVAL=2;UNTIL=20261231\r\nEXDATE;VALUE=DATE:20261116\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:b\r\nSUMMARY:Gelber Sack\\, Papier\r\nDTSTART;VALUE=DATE:20261112\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:c\r\nSUMMARY:Monthly by day\r\nDTSTART;VALUE=DATE:20261103\r\nRRULE:FREQ=MONTHLY;BYDAY=1TU\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

func TestParseAndExpand(t *testing.T) {
	events, err := Parse(wasteICS, time.UTC)
	if err != nil || len(events) != 3 {
		t.Fatalf("%+v %v", events, err)
	}
	if events[1].Summary != "Gelber Sack, Papier" || !events[1].AllDay || events[0].RRule == "" || len(events[0].ExDates) != 1 {
		t.Fatalf("parse: %+v", events)
	}
	from, to := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
	bio := Expand(events[0], from, to)
	var dates []string
	for _, e := range bio {
		dates = append(dates, e.Start.Format("01-02"))
	}
	if got := strings.Join(dates, ","); got != "11-02,11-30,12-14,12-28" {
		t.Fatalf("expansion: %s", got)
	}
	if n := len(Expand(events[2], from, to)); n != 1 {
		t.Fatalf("BY rules must not be guessed: %d", n)
	}
	if n := len(Expand(events[1], from, to)); n != 1 {
		t.Fatal("single event lost")
	}
	count := Event{Start: from, End: from.Add(time.Hour), RRule: "FREQ=DAILY;COUNT=3"}
	if n := len(Expand(count, from, to)); n != 3 {
		t.Fatalf("COUNT: %d", n)
	}
}

func TestExpandWeeklyOnSeveralDays(t *testing.T) {
	// Wednesday 4 Nov 2026, 08:00 UTC: a lecture on Mondays and Wednesdays.
	start := time.Date(2026, 11, 4, 8, 0, 0, 0, time.UTC)
	e := Event{Start: start, End: start.Add(90 * time.Minute), RRule: "FREQ=WEEKLY;BYDAY=MO,WE;COUNT=5",
		ExDates: []time.Time{time.Date(2026, 11, 11, 8, 0, 0, 0, time.UTC)}}
	var got []string
	for _, occ := range Expand(e, start.AddDate(0, 0, -7), start.AddDate(0, 2, 0)) {
		got = append(got, occ.Start.Format("Mon 01-02"))
	}
	// DTSTART counts as the first occurrence, the Monday before it does not;
	// the excluded date still counts towards COUNT.
	if s := strings.Join(got, ","); s != "Wed 11-04,Mon 11-09,Mon 11-16,Wed 11-18" {
		t.Fatalf("expansion: %s", s)
	}
	biweekly := Event{Start: start, End: start.Add(time.Hour), RRule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=WE,FR;UNTIL=20261121T000000Z"}
	got = nil
	for _, occ := range Expand(biweekly, start, start.AddDate(0, 1, 0)) {
		got = append(got, occ.Start.Format("01-02"))
	}
	if s := strings.Join(got, ","); s != "11-04,11-06,11-18,11-20" {
		t.Fatalf("interval and until: %s", s)
	}
}

func TestWriteHelpers(t *testing.T) {
	if EscapeText("a;b,c\\d\ne") != `a\;b\,c\\d\ne` {
		t.Fatal(EscapeText("a;b,c\\d\ne"))
	}
	long := "SUMMARY:" + strings.Repeat("ä", 60)
	for _, l := range strings.Split(strings.TrimSuffix(FoldLine(long), "\r\n"), "\r\n") {
		if len(l) > 75 {
			t.Fatalf("line too long: %d", len(l))
		}
	}
}
