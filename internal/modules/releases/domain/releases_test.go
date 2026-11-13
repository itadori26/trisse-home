package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

const mbid = "a74b1b7f-71a5-4011-9441-d0b5e4122711"

func TestNewArtist(t *testing.T) {
	if a, err := NewArtist(mbid, " Radiohead "); err != nil || a.Name != "Radiohead" {
		t.Fatalf("%+v %v", a, err)
	}
	for _, c := range [][2]string{{"x", "A"}, {mbid, " "}} {
		if _, err := NewArtist(c[0], c[1]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%v accepted", c)
		}
	}
}

func TestRecent(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	list := []Release{
		{Title: "old", Date: now.AddDate(0, 0, -40)},
		{Title: "b", Date: now.AddDate(0, 0, -3)},
		{Title: "future", Date: now.AddDate(0, 0, 3)},
		{Title: "a", Date: now.AddDate(0, 0, -3)},
		{Title: "new", Date: now.AddDate(0, 0, -1)},
	}
	got := Recent(list, 4, now)
	if len(got) != 3 || got[0].Title != "new" || got[1].Title != "a" || got[2].Title != "b" {
		t.Fatalf("newest first within four weeks: %+v", got)
	}
	if len(Recent(list, 0, now)) != 3 || ClampWeeks(99) != MaxWeeks {
		t.Fatal("weeks are bounded")
	}
}

func TestParseDate(t *testing.T) {
	for in, want := range map[string]string{"2026-11-07": "2026-11-07", "2026-11": "2026-11-01", "2026": "2026-01-01"} {
		if d, ok := ParseDate(in); !ok || d.Format(time.DateOnly) != want {
			t.Errorf("ParseDate(%q) = %v %v", in, d, ok)
		}
	}
	if _, ok := ParseDate(""); ok {
		t.Fatal("empty date")
	}
}
