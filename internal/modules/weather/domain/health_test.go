package domain

import "testing"

func TestPollen(t *testing.T) {
	for s, want := range map[string]PollenLevel{"0": 0, "0-1": 1, "1": 2, "1-2": 3, "2": 4, "2-3": 5, "3": 6, "-1": NoPollenData, "": NoPollenData} {
		if got := ParsePollenLevel(s); got != want {
			t.Errorf("%q: %d, want %d", s, got, want)
		}
	}
	if PollenLevel(3).String() != "1-2" || NoPollenData.String() != "" {
		t.Fatal("String")
	}
	for l, want := range map[PollenLevel]string{-1: "unknown", 0: "none", 1: "low", 2: "low", 3: "moderate", 4: "moderate", 5: "high", 6: "high"} {
		if got := l.Burden(); got != want {
			t.Errorf("Burden(%d) = %s, want %s", l, got, want)
		}
	}
	if p, ok := ParsePollenType("Graeser"); !ok || p != Grasses {
		t.Fatal("Graeser")
	}
	if p, ok := ParsePollenType("birch"); !ok || p != Birch {
		t.Fatal("birch")
	}
	if _, ok := ParsePollenType("Pilze"); ok {
		t.Fatal("unknown type")
	}
	f := PollenForecast{Days: []PollenDay{{Type: Mugwort}, {Type: Hazel}, {Type: Birch}}}
	if got := f.Of(nil); len(got) != 3 || got[0].Type != Hazel || got[2].Type != Mugwort {
		t.Fatalf("season order: %+v", got)
	}
	if got := f.Of([]PollenType{Birch}); len(got) != 1 || got[0].Type != Birch {
		t.Fatalf("filter: %+v", got)
	}
}

func TestUVLevel(t *testing.T) {
	for i, want := range map[int]string{-1: "unknown", 0: "low", 2: "low", 3: "moderate", 5: "moderate", 6: "high", 7: "high", 8: "very_high", 10: "very_high", 11: "extreme"} {
		if got := UVLevel(i); got != want {
			t.Errorf("UVLevel(%d) = %s, want %s", i, got, want)
		}
	}
}
