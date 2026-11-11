package domain

import (
	"testing"
	"time"
)

func TestNewPolicy(t *testing.T) {
	for _, tc := range []struct {
		interval  time.Duration
		retention int
		ok        bool
	}{{24 * time.Hour, 7, true}, {time.Minute, 7, false}, {time.Hour, 0, false}, {time.Hour, 366, false}} {
		if _, err := NewPolicy(tc.interval, tc.retention); (err == nil) != tc.ok {
			t.Errorf("NewPolicy(%v,%d) err=%v", tc.interval, tc.retention, err)
		}
	}
}

func TestNamesAndPruning(t *testing.T) {
	t0 := time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)
	name := FileName(t0)
	if name != "trisse-home-20261111-030000.db" || !ValidName(name) {
		t.Fatalf("unexpected name %q", name)
	}
	for _, bad := range []string{"../etc/passwd", "trisse-home-20261111-030000.db/../x", "other.db", ""} {
		if ValidName(bad) {
			t.Errorf("%q accepted", bad)
		}
	}
	p, _ := NewPolicy(24*time.Hour, 2)
	files := []File{{Name: "a", CreatedAt: t0}, {Name: "c", CreatedAt: t0.Add(48 * time.Hour)}, {Name: "b", CreatedAt: t0.Add(24 * time.Hour)}}
	prune := p.ToPrune(files)
	if len(prune) != 1 || prune[0].Name != "a" {
		t.Fatalf("prune %+v", prune)
	}
	if !p.Due(time.Time{}, t0) || p.Due(t0, t0.Add(time.Hour)) || !p.Due(t0, t0.Add(24*time.Hour-30*time.Second)) {
		t.Fatal("Due wrong")
	}
}
