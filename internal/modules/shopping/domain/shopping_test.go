package domain

import (
	"testing"
	"time"
)

func TestDomain(t *testing.T) {
	if !ValidItemID("abc-123@trisse-home.ics") || ValidItemID("../x.ics") || ValidItemID("a/b.ics") || ValidItemID("x.txt") {
		t.Fatal("item ID validation wrong")
	}
	if s, err := NormaliseSummary("  Milk   2 l "); err != nil || s != "Milk 2 l" {
		t.Fatalf("%q %v", s, err)
	}
	if _, err := NormaliseSummary("   "); err == nil {
		t.Fatal("empty item accepted")
	}
	t0 := time.Unix(1_800_000_000, 0)
	items := []Item{
		{ID: "b", Summary: "Bread", Created: t0.Add(time.Minute)},
		{ID: "d1", Summary: "Done old", Done: true, Completed: t0},
		{ID: "a", Summary: "Apples", Created: t0},
		{ID: "d2", Summary: "Done new", Done: true, Completed: t0.Add(time.Hour)},
	}
	got := ForDisplay(items)
	if got[0].ID != "a" || got[1].ID != "b" || got[2].ID != "d2" || got[3].ID != "d1" {
		t.Fatalf("order %+v", got)
	}
	if s, err := NewSettings("https://c/remote.php/dav/calendars/me/shopping", "me", "pw"); err != nil || s.ListURL[len(s.ListURL)-1] != '/' {
		t.Fatalf("%+v %v", s, err)
	}
}
