package domain

import (
	"testing"
	"time"
)

func TestRelevantWarnings(t *testing.T) {
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	ws := []Warning{
		{ID: "1", Severity: SeverityMinor, Onset: now, Expires: now.Add(time.Hour)},
		{ID: "2", Severity: SeveritySevere, Onset: now.Add(time.Hour), Expires: now.Add(3 * time.Hour)},
		{ID: "2", Severity: SeveritySevere, Onset: now.Add(time.Hour), Expires: now.Add(3 * time.Hour)},
		{ID: "3", Severity: SeverityExtreme, Onset: now.Add(-3 * time.Hour), Expires: now.Add(-time.Hour)},
	}
	got := RelevantWarnings(ws, now)
	if len(got) != 2 || got[0].ID != "2" || got[1].ID != "1" {
		t.Fatalf("%+v", got)
	}
	if ParseSeverity("Moderate") != SeverityModerate || ParseSeverity("?") != SeverityMinor {
		t.Fatal("severity parsing wrong")
	}
	oldenburg, _ := NewLocation(53.14, 8.21)
	paris, _ := NewLocation(48.86, 2.35)
	if !oldenburg.InGermany() || paris.InGermany() {
		t.Fatal("coverage check wrong")
	}
	if b := oldenburg.RadarBBox(); b.MinLat >= 53.14 || b.MaxLon <= 8.21 {
		t.Fatalf("bbox %+v", b)
	}
}
