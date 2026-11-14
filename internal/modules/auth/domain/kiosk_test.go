package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestKiosk(t *testing.T) {
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	k, err := NewKiosk("k1", "flur-tablet", []string{"t1", "t1", ""}, now)
	if err != nil || k.Role() != RoleKiosk || len(k.Tabs()) != 1 || k.CanSee("t2") || !k.CanSee("t1") {
		t.Fatalf("%+v %v", k, err)
	}
	if _, err := NewKiosk("k1", "Flur Tablet", []string{"t1"}, now); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("name")
	}
	if _, err := NewKiosk("k1", "flur", nil, now); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("a display needs a tab")
	}
	if o, err := NewKioskOptions(60, "22:00", "06:30"); err != nil || o.NightTo != "06:30" {
		t.Fatal(o, err)
	}
	for name, err := range map[string]error{
		"rotate":   func() error { _, e := NewKioskOptions(5, "", ""); return e }(),
		"only one": func() error { _, e := NewKioskOptions(0, "22:00", ""); return e }(),
		"format":   func() error { _, e := NewKioskOptions(0, "25:00", "06:00"); return e }(),
		"same":     func() error { _, e := NewKioskOptions(0, "06:00", "06:00"); return e }(),
	} {
		if !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestKioskMayAccess(t *testing.T) {
	mods := map[string]bool{"weather": true}
	for _, c := range []struct {
		method, path string
		want         bool
	}{
		{"GET", "/api/dashboard", true},
		{"GET", "/api/weather/forecast", true},
		{"GET", "/api/auth/kiosk", true},
		{"POST", "/api/auth/logout", true},
		{"GET", "/api/todos", false},
		{"GET", "/api/weather/settings", false},
		{"PUT", "/api/auth/password", false},
		{"POST", "/api/auth/totp/setup", false},
		{"POST", "/api/auth/passkeys/register/begin", false},
		{"GET", "/api/auth/sessions", false},
		{"POST", "/api/weather/x", false},
		{"GET", "/api/auth/kiosks", false},
	} {
		if got := KioskMayAccess(c.method, c.path, mods); got != c.want {
			t.Errorf("%s %s = %v", c.method, c.path, got)
		}
	}
}
