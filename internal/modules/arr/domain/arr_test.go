package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestNewInstance(t *testing.T) {
	in, err := NewInstance("1", "sonarr", "", " http://sonarr:8989/ ", " key ")
	if err != nil || in.Name != "Sonarr" || in.URL != "http://sonarr:8989" || in.APIKey != "key" {
		t.Fatalf("unexpected %+v %v", in, err)
	}
	for name, c := range map[string][4]string{
		"unknown app": {"lidarr", "", "http://x", "k"},
		"bad url":     {"radarr", "", "radarr:7878", "k"},
		"no key":      {"radarr", "", "http://x", " "},
		"long name":   {"radarr", "0123456789012345678901234567890123456789x", "http://x", "k"},
	} {
		if _, err := NewInstance("x", c[0], c[1], c[2], c[3]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: expected a validation error, got %v", name, err)
		}
	}
	if !AppSonarr.HasLibrary() || AppProwlarr.HasLibrary() {
		t.Fatal("HasLibrary")
	}
}

func TestMerge(t *testing.T) {
	day := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	o := Merge([]string{"Sonarr", "Radarr"}, []Snapshot{
		{Upcoming: []Release{{Title: "late", Date: day.Add(48 * time.Hour)}}, Health: []HealthIssue{{Severity: SeverityWarning}}, Missing: 3},
		{Upcoming: []Release{{Title: "early", Date: day}}, Health: []HealthIssue{{Severity: SeverityError}}},
	})
	if o.Upcoming[0].Title != "early" || o.Health[0].Severity != SeverityError {
		t.Fatalf("sorted by date and severity: %+v", o)
	}
	if o.Missing["Sonarr"] != 3 || len(o.Missing) != 1 {
		t.Fatalf("missing: %v", o.Missing)
	}
}
