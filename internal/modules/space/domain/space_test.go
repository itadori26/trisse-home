package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestDomain(t *testing.T) {
	for in, want := range map[string]LaunchStatus{"Go": StatusGo, "TBD": StatusTBD, "In Flight": StatusInFlight, "Partial Failure": StatusPartial, "???": StatusUnknown} {
		if got := ParseLaunchStatus(in); got != want {
			t.Errorf("%s: %s", in, got)
		}
	}
	if Truncate("abcdef", 4) != "abc…" || Truncate(" ab ", 4) != "ab" {
		t.Fatal("Truncate")
	}
	if (Picture{MediaType: "video", URL: "x"}).IsImage() || !(Picture{MediaType: "image", URL: "x"}).IsImage() {
		t.Fatal("IsImage")
	}
	stored := Settings{NASAKey: "abc"}
	if s, _ := NewSettings(true, true, "", true, stored); s.NASAKey != "abc" || !s.HasKey() {
		t.Fatalf("keep: %+v", s)
	}
	if s, _ := NewSettings(true, false, "", false, stored); s.HasKey() {
		t.Fatal("remove key")
	}
	if _, err := NewSettings(true, true, "a&b", false, stored); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
}
