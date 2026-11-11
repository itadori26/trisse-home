package domain

import (
	"testing"
	"time"
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
	if err != nil || s.URL != "https://example.org/abfall.ics" || s.Rules[0].Match != "bio" {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := NewSettings("https://x", []Rule{{Match: "x", Color: "green"}}); err == nil {
		t.Fatal("invalid color accepted")
	}
	if _, err := NewSettings("ftp://x", nil); err == nil {
		t.Fatal("invalid URL accepted")
	}
}
