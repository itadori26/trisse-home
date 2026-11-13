package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestNewSettings(t *testing.T) {
	s, err := NewSettings("jellyfin", " http://jf:8096/ ", " key ")
	if err != nil || s.URL != "http://jf:8096" || s.Token != "key" || s.Provider != ProviderJellyfin {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	for _, c := range [][3]string{{"emby", "http://x", "k"}, {"plex", "x", "k"}, {"plex", "http://x", " "}} {
		if _, err := NewSettings(c[0], c[1], c[2]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%v: expected a validation error, got %v", c, err)
		}
	}
}

func TestValidItemID(t *testing.T) {
	for id, ok := range map[string]bool{"4e7b3a2c9d1f": true, "12345": true, "a-b": true, "../etc": false, "a/b": false, "": false} {
		if ValidItemID(id) != ok {
			t.Errorf("ValidItemID(%q) != %v", id, ok)
		}
	}
}

func TestProgress(t *testing.T) {
	if p := (Session{Position: 30 * time.Minute, Duration: time.Hour}).Progress(); p != 50 {
		t.Fatalf("progress %v", p)
	}
	if (Session{}).Progress() != 0 || (Session{Position: 2 * time.Hour, Duration: time.Hour}).Progress() != 100 {
		t.Fatal("progress bounds")
	}
}
