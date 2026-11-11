package domain

import (
	"testing"
	"time"
)

func TestDepartures(t *testing.T) {
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	ds := []Departure{
		{Line: "gone", Planned: now.Add(-5 * time.Minute)},
		{Line: "late", Planned: now.Add(-2 * time.Minute), Expected: now.Add(6 * time.Minute), RealTime: true},
		{Line: "soon", Planned: now.Add(3 * time.Minute)},
		{Line: "later", Planned: now.Add(10 * time.Minute)},
	}
	got := Upcoming(ds, now, 2)
	if len(got) != 2 || got[0].Line != "soon" || got[1].Line != "late" || got[1].DelayMinutes() != 8 {
		t.Fatalf("%+v", got)
	}
	if (Departure{Planned: now, Expected: now.Add(time.Hour)}).DelayMinutes() != 0 {
		t.Fatal("delay without real-time data")
	}
	if ValidStopID("") || ValidStopID("a\nb") || !ValidStopID("de-DELFI_de:03403:16750") {
		t.Fatal("stop ID validation wrong")
	}
}
