package domain

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestClassify(t *testing.T) {
	if s := Classify("Restmüll 2-wöchentlich", nil); s.Icon != "🗑️" || s.Label != "Restmüll 2-wöchentlich" {
		t.Fatalf("%+v", s)
	}
	if s := Classify("Gelber Sack", nil); s.Icon != "♻️" {
		t.Fatalf("%+v", s)
	}
	if s := Classify("Grünschnitt", nil); s.Icon != "🌳" {
		t.Fatal("garden waste must win over generic keywords")
	}
	rules := []Rule{{Match: "gelb", Label: "Yellow bag", Color: "#ffff00"}}
	if s := Classify("Gelbe Tonne", rules); s.Label != "Yellow bag" || s.Color != "#ffff00" || s.Icon != "♻️" {
		t.Fatalf("custom rule: %+v", s)
	}
}

func TestNextCollections(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC) }
	cs := []Collection{
		{d(10), "Biotonne"}, {d(12), "Biotonne"}, {d(26), "Biotonne"}, {d(12), "Biotonne"},
		{d(20), "Papier"}, {d(13), "Restmüll"}, {d(27), "Biotonne"}, {d(28), "Biotonne"},
	}
	today := time.Date(2026, 11, 11, 21, 30, 0, 0, time.FixedZone("CET", 3600))
	got := NextCollections(cs, nil, today)
	if len(got) != 3 || got[0].Style.Label != "Biotonne" || got[0].DaysUntil != 1 || len(got[0].Dates) != 3 {
		t.Fatalf("%+v", got)
	}
	if got[1].Style.Label != "Restmüll" || got[2].DaysUntil != 9 {
		t.Fatalf("order: %+v", got)
	}
}

func TestSettings(t *testing.T) {
	s, err := NewSettings("webcal://example.org/abfall.ics", []Rule{{Match: " bio ", Color: "#00ff00"}})
	if err != nil || s.Source.URL != "https://example.org/abfall.ics" || s.Rules[0].Match != "bio" {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := NewSettings("https://x", []Rule{{Match: "x", Color: "green"}}); err == nil {
		t.Fatal("invalid color accepted")
	}
	if _, err := NewSettings("ftp://x", nil); err == nil {
		t.Fatal("invalid URL accepted")
	}
}

func TestFileSource(t *testing.T) {
	day := func(m time.Month, d int) time.Time { return time.Date(2027, m, d, 0, 0, 0, 0, time.UTC) }
	if _, err := NewFileSource("a.ics", day(1, 1), nil); err == nil {
		t.Fatal("a file without dates must be rejected")
	}
	src, err := NewFileSource("../../etc/\x00plan.ics", day(1, 1), []Collection{{Date: day(3, 2)}, {Date: day(1, 5)}, {Date: day(12, 28)}})
	if err != nil {
		t.Fatal(err)
	}
	f := src.File
	if src.Kind != SourceFile || f.Name != "plan.ics" || f.Events != 3 || !f.From.Equal(day(1, 5)) || !f.To.Equal(day(12, 28)) {
		t.Fatalf("%+v", src)
	}
	if long, _ := NewFileSource(strings.Repeat("ä", 150)+".ics", day(1, 1), []Collection{{Date: day(1, 5)}}); utf8.RuneCountInString(long.File.Name) != 100 {
		t.Fatalf("name not cut: %q", long.File.Name)
	}
	// The hint starts 30 days before the last date and stays after it.
	at := func(m time.Month, d int) time.Time { return time.Date(2027, m, d, 18, 0, 0, 0, time.UTC) }
	for now, want := range map[time.Time]bool{at(11, 27): false, at(11, 28): true, at(12, 28): true, at(12, 30): true} {
		if got := src.ExpiresWithin(now, ExpiryWarning); got != want {
			t.Errorf("ExpiresWithin(%v) = %v, want %v", now, got, want)
		}
	}
	url, _ := NewURLSource("webcal://example.org/a.ics")
	if url.Kind != SourceURL || url.URL != "https://example.org/a.ics" || url.ExpiresWithin(at(12, 30), ExpiryWarning) {
		t.Fatalf("%+v", url)
	}
	if (Source{}).Configured() || !url.Configured() {
		t.Fatal("Configured")
	}
}
