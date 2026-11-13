package domain

import (
	"errors"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestEaster(t *testing.T) {
	for year, want := range map[int]string{2019: "2019-04-21", 2024: "2024-03-31", 2025: "2025-04-20", 2026: "2026-04-05", 2027: "2027-03-28", 2038: "2038-04-25"} {
		if got := Easter(year).Format(time.DateOnly); got != want {
			t.Errorf("Easter(%d) = %s, want %s", year, got, want)
		}
	}
}

func keys(hs []Holiday) []string {
	out := make([]string, 0, len(hs))
	for _, h := range hs {
		out = append(out, h.Date.Format("01-02")+" "+h.Key)
	}
	return out
}

func TestPublicHolidays(t *testing.T) {
	ni := keys(PublicHolidays(2026, "NI"))
	want := []string{"01-01 new_year", "04-03 good_friday", "04-06 easter_monday", "05-01 labour_day", "05-14 ascension",
		"05-25 whit_monday", "10-03 german_unity", "10-31 reformation", "12-25 christmas_day", "12-26 boxing_day"}
	if !slices.Equal(ni, want) {
		t.Fatalf("NI 2026:\n got %v\nwant %v", ni, want)
	}
	for state, extra := range map[State][]string{
		"BY": {"01-06 epiphany", "06-04 corpus_christi", "11-01 all_saints"},
		"BE": {"03-08 womens_day"},
		"BB": {"04-05 easter_sunday", "05-24 whit_sunday", "10-31 reformation"},
		"SN": {"10-31 reformation", "11-18 repentance"},
		"SL": {"06-04 corpus_christi", "08-15 assumption", "11-01 all_saints"},
		"TH": {"09-20 childrens_day", "10-31 reformation"},
	} {
		got := keys(PublicHolidays(2026, state))
		for _, e := range extra {
			if !slices.Contains(got, e) {
				t.Errorf("%s 2026 misses %s: %v", state, e, got)
			}
		}
	}
	if len(PublicHolidays(2026, "BY")) != 12 || len(PublicHolidays(2026, "HB")) != 10 || len(PublicHolidays(2026, "BW")) != 12 {
		t.Fatalf("counts BY=%d HB=%d BW=%d", len(PublicHolidays(2026, "BY")), len(PublicHolidays(2026, "HB")), len(PublicHolidays(2026, "BW")))
	}
	// Rules with a start year: Reformation Day in Lower Saxony since 2018, Women's Day in MV since 2023.
	if slices.Contains(keys(PublicHolidays(2016, "NI")), "10-31 reformation") || slices.Contains(keys(PublicHolidays(2022, "MV")), "03-08 womens_day") {
		t.Fatal("holiday before its first year")
	}
	if !slices.Contains(keys(PublicHolidays(2016, "SN")), "10-31 reformation") {
		t.Fatal("Saxony has always had Reformation Day")
	}
	// Buß- und Bettag is the Wednesday before 23 November.
	for year, want := range map[int]string{2024: "11-20", 2025: "11-19", 2027: "11-17", 2028: "11-22"} {
		hs := PublicHolidays(year, "SN")
		i := slices.IndexFunc(hs, func(h Holiday) bool { return h.Key == "repentance" })
		if i < 0 || hs[i].Date.Format("01-02") != want || hs[i].Date.Weekday() != time.Wednesday {
			t.Errorf("repentance %d: %v", year, hs)
		}
	}
}

func TestUpcoming(t *testing.T) {
	today := time.Date(2026, 12, 21, 15, 0, 0, 0, time.Local)
	school := []SchoolHoliday{
		{Name: "Weihnachtsferien", Start: day(2026, 12, 21), End: day(2027, 1, 5)},
		{Name: "Herbstferien", Start: day(2026, 10, 12), End: day(2026, 10, 24)},
		{Name: "Winterferien", Start: day(2027, 2, 1), End: day(2027, 2, 2)},
	}
	got := Upcoming(today, "NI", school, 5)
	var labels []string
	for _, e := range got {
		labels = append(labels, e.Start.Format("01-02")+" "+e.Key+e.Name)
	}
	want := []string{"12-21 Weihnachtsferien", "12-25 christmas_day", "12-26 boxing_day", "01-01 new_year", "02-01 Winterferien"}
	if !slices.Equal(labels, want) {
		t.Fatalf("\n got %v\nwant %v", labels, want)
	}
	if !got[0].Current || got[0].DaysUntil != 0 || got[1].Current || got[1].DaysUntil != 4 {
		t.Fatalf("%+v", got[:2])
	}
	// Today is a holiday: current, 0 days.
	if e := Upcoming(day(2026, 12, 25), "NI", nil, 1); e[0].Key != "christmas_day" || !e[0].Current {
		t.Fatalf("%+v", e)
	}
	if n := len(Upcoming(day(2026, 1, 1), "NI", nil, 99)); n != MaxEntries {
		t.Fatalf("limit: %d", n)
	}
}

func TestStatesAndSettings(t *testing.T) {
	if ParseState("by") != "BY" || ParseState("XX") != DefaultState || ParseState("") != "NI" || len(States) != 16 {
		t.Fatal("ParseState")
	}
	s, err := NewSettings(" webcal://www.example.org/ferien/ni.ics ")
	if err != nil || s.SchoolURL != "https://www.example.org/ferien/ni.ics" {
		t.Fatalf("%+v %v", s, err)
	}
	if s, err := NewSettings(""); err != nil || s.SchoolURL != "" {
		t.Fatalf("empty: %+v %v", s, err)
	}
	for _, bad := range []string{"ftp://x/y.ics", "http://169.254.169.254/latest", "https://user:pw@x/y.ics"} {
		if _, err := NewSettings(bad); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", bad, err)
		}
	}
}
