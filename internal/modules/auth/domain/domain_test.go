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
	s, err := NewSession(HashToken("raw"), "csrf", now, ttl)
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
	if _, err := NewSession("", "csrf", time.Now(), time.Hour); err == nil {
		t.Fatal("expected error for empty token hash")
	}
	if _, err := NewSession("h", "csrf", time.Now(), 0); err == nil {
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
