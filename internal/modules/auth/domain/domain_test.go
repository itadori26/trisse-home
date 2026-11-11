package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"too short", "short", true},
		{"exactly min", strings.Repeat("a", MinPasswordLength), false},
		{"too long", strings.Repeat("a", MaxPasswordLength+1), true},
		{"whitespace", strings.Repeat(" ", 20), true},
		{"multibyte counts runes", strings.Repeat("ä", MinPasswordLength), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPassword(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, kernel.ErrInvalid) {
				t.Fatalf("expected ErrInvalid, got %v", err)
			}
		})
	}
}

func TestSessionLifecycle(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl := 24 * time.Hour
	s, err := NewSession(OwnerID, HashToken("raw"), "pub", "csrf", NewClient("ua", "1.2.3.4"), now, ttl)
	if err != nil {
		t.Fatal(err)
	}
	if s.IsExpired(now.Add(ttl - time.Second)) {
		t.Fatal("expired too early")
	}
	if !s.IsExpired(now.Add(ttl)) {
		t.Fatal("not expired at expiry")
	}
	if s.Refresh(now.Add(time.Minute), ttl) {
		t.Fatal("refreshed within refresh interval")
	}
	later := now.Add(2 * time.Hour)
	if !s.Refresh(later, ttl) || !s.ExpiresAt().Equal(later.Add(ttl)) {
		t.Fatal("refresh did not extend expiry")
	}
}

func TestNewSessionRejectsInvalidInput(t *testing.T) {
	if _, err := NewSession(OwnerID, "", "pub", "csrf", Client{}, time.Now(), time.Hour); err == nil {
		t.Fatal("expected error for empty token hash")
	}
	if _, err := NewSession(OwnerID, "h", "pub", "csrf", Client{}, time.Now(), 0); err == nil {
		t.Fatal("expected error for zero ttl")
	}
}

func TestHashTokenIsDeterministic(t *testing.T) {
	a1, a2, b := HashToken("a"), HashToken("a"), HashToken("b")
	if a1 != a2 || a1 == b {
		t.Fatal("hash not deterministic or colliding")
	}
	if len(HashToken("a")) != 64 {
		t.Fatal("expected hex sha256")
	}
}

func TestAccount(t *testing.T) {
	now := time.Now()
	if _, err := NewAccount("", now); err == nil {
		t.Fatal("expected error for empty hash")
	}
	a, _ := NewAccount("h1", now)
	if err := a.ChangePasswordHash("h2", now.Add(time.Hour)); err != nil || a.PasswordHash() != "h2" {
		t.Fatal("password hash not changed")
	}
}

func TestNewClientTruncatesUserAgent(t *testing.T) {
	c := NewClient(strings.Repeat("x", 1000), "10.0.0.1")
	if len(c.UserAgent) != maxUserAgentLength || c.IP != "10.0.0.1" {
		t.Fatalf("unexpected %d %q", len(c.UserAgent), c.IP)
	}
}

func TestGuestAccount(t *testing.T) {
	now := time.Unix(0, 0)
	g, err := NewGuest("g1", "anna", "hash", []string{"t1", "t1", "", "t2"}, now)
	if err != nil || g.IsOwner() || len(g.Tabs()) != 2 || !g.CanSee("t2") || g.CanSee("t3") {
		t.Fatalf("%+v %v", g, err)
	}
	for _, bad := range []string{"", "a", "Anna", "owner", "has space", "x/y"} {
		if _, err := NewGuest("g", bad, "hash", nil, now); err == nil {
			t.Errorf("name %q accepted", bad)
		}
	}
	_ = g.SetGuestAccess([]string{"t3"}, true, now)
	if !g.Disabled() || !g.CanSee("t3") || g.CanSee("t1") {
		t.Fatal("access not updated")
	}
	owner, _ := NewAccount("hash", now)
	if !owner.CanSee("anything") || owner.SetGuestAccess(nil, true, now) == nil {
		t.Fatal("owner sees everything and is not a guest")
	}
}
