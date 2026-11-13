package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(ProviderAdGuard, " http://10.0.0.2:3000/ ", "admin", "secret")
	if err != nil || s.URL != "http://10.0.0.2:3000" || s.Provider != ProviderAdGuard {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	for _, tc := range [][3]string{{"10.0.0.2", "a", "b"}, {"http://x", "", "b"}, {"http://x", "a", ""}} {
		if _, err := NewSettings(ProviderAdGuard, tc[0], tc[1], tc[2]); err == nil {
			t.Errorf("expected error for %v", tc)
		}
	}
	// Pi-hole needs no username, and no password when it has none.
	pi, err := NewSettings(ProviderPiHole, "http://pi.hole", "ignored", "")
	if err != nil || pi.Username != "" || pi.Password != "" {
		t.Fatalf("pi-hole: %+v %v", pi, err)
	}
	if _, err := NewSettings("bind", "http://x", "a", "b"); err == nil {
		t.Fatal("unknown provider accepted")
	}
}

func TestParseProvider(t *testing.T) {
	for raw, want := range map[string]Provider{"": ProviderAdGuard, "adguard": ProviderAdGuard, "pihole": ProviderPiHole, "\tpihole\n": ProviderPiHole} {
		if got, err := ParseProvider(raw); err != nil || got != want {
			t.Errorf("ParseProvider(%q) = %q, %v", raw, got, err)
		}
	}
	if _, err := ParseProvider("bind"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown provider: %v", err)
	}
}

func TestBlockedPercent(t *testing.T) {
	if (Stats{}).BlockedPercent() != 0 || (Stats{Queries: 200, Blocked: 50}).BlockedPercent() != 25 {
		t.Fatal("percent wrong")
	}
}
