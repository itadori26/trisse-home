package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestItems(t *testing.T) {
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	link, err := NewText("a", " https://example.org/x?y=1 ", 0, now)
	if err != nil || link.Kind != KindLink || !link.ExpiresAt.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("%+v %v", link, err)
	}
	if txt, _ := NewText("b", "Treffpunkt https://example.org um 8", time.Hour, now); txt.Kind != KindText {
		t.Fatalf("text with a link stays text: %+v", txt)
	}
	for name, err := range map[string]error{
		"empty":   func() error { _, e := NewText("c", "  ", 0, now); return e }(),
		"large":   func() error { _, e := NewText("c", strings.Repeat("x", MaxTextBytes+1), 0, now); return e }(),
		"expiry":  func() error { _, e := NewText("c", "x", 2*time.Hour, now); return e }(),
		"file":    func() error { _, e := NewFile("c", "a.bin", "", MaxFileBytes+1, 0, now); return e }(),
		"nothing": func() error { _, e := NewFile("c", "a.bin", "", 0, 0, now); return e }(),
	} {
		if !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if !link.Expired(now.Add(24*time.Hour)) || link.Expired(now.Add(time.Hour)) {
		t.Fatal("expiry")
	}
	for in, want := range map[string]string{`C:\Users\me\Bild 1.png`: "Bild 1.png", "../../etc/passwd": "passwd", `a"b.txt`: "ab.txt", "": "file", "..": "file"} {
		if got := CleanFileName(in); got != want {
			t.Errorf("CleanFileName(%q) = %q", in, got)
		}
	}
	for ct, want := range map[string]string{"image/png": "image/png", "text/html": "application/octet-stream", "image/svg+xml": "application/octet-stream", "": "application/octet-stream"} {
		if got := SafeInlineType(ct); got != want {
			t.Errorf("SafeInlineType(%q) = %q", ct, got)
		}
	}
}

func TestQuota(t *testing.T) {
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	big := []Item{{Size: QuotaBytes - 10, ExpiresAt: now.Add(time.Hour)}}
	if err := CheckQuota(big, now, 11); !errors.Is(err, ErrQuotaFull) {
		t.Fatal("size quota")
	}
	if err := CheckQuota(big, now.Add(2*time.Hour), 11); err != nil {
		t.Fatal("expired items do not count")
	}
	many := make([]Item, MaxItems)
	for i := range many {
		many[i].ExpiresAt = now.Add(time.Hour)
	}
	if err := CheckQuota(many, now, 1); !errors.Is(err, ErrQuotaFull) {
		t.Fatal("item count")
	}
}
