package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" 76561197960287930 ", "de", 25, true)
	if err != nil || s.Country != "DE" || !s.Enabled() {
		t.Fatalf("%+v %v", s, err)
	}
	if s, _ := NewSettings("", "AT", 0, false); s.Enabled() {
		t.Fatal("nothing is enabled without Steam ID and Epic")
	}
	for name, c := range map[string]struct {
		id, country string
		min         int
	}{
		"vanity name": {"gaben", "DE", 0}, "country": {"", "Germany", 0}, "discount": {"", "DE", 101},
	} {
		if _, err := NewSettings(c.id, c.country, c.min, false); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestFilterDeals(t *testing.T) {
	got := FilterDeals([]Deal{{Name: "b", DiscountPct: 50}, {Name: "none"}, {Name: "small", DiscountPct: 10}, {Name: "a", DiscountPct: 50}, {Name: "max", DiscountPct: 90}}, 20)
	if len(got) != 3 || got[0].Name != "max" || got[1].Name != "a" || got[2].Name != "b" {
		t.Fatalf("biggest discount first, at least the minimum: %+v", got)
	}
}

func TestFreeGameCurrent(t *testing.T) {
	now := time.Date(2026, 11, 12, 16, 0, 0, 0, time.UTC)
	g := FreeGame{Start: now.Add(-time.Hour), End: now.Add(time.Hour)}
	if !g.Current(now) || g.Current(now.Add(time.Hour)) {
		t.Fatal("current window")
	}
}
