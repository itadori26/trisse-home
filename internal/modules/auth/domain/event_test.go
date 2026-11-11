package domain

import (
	"strings"
	"testing"
	"time"
)

func TestLoginThrottleBackoff(t *testing.T) {
	th := DefaultLoginThrottle
	want := []time.Duration{15 * time.Minute, 30 * time.Minute, time.Hour, 2 * time.Hour}
	for n, w := range want {
		if got := th.Lockout(n); got != w {
			t.Errorf("lockout %d: %v, want %v", n, got, w)
		}
	}
	if th.Lockout(50) != 24*time.Hour {
		t.Fatal("backoff must be capped")
	}
}

func TestNewSecurityEvent(t *testing.T) {
	ev, err := NewSecurityEvent("1", time.Unix(0, 0), EventLoginFailed, "owner", Client{IP: "10.0.0.2"}, strings.Repeat("x", 500))
	if err != nil || len(ev.Detail) != maxEventDetail {
		t.Fatalf("%v %d", err, len(ev.Detail))
	}
	if _, err := NewSecurityEvent("", time.Unix(0, 0), EventLoginFailed, "", Client{}, ""); err == nil {
		t.Fatal("missing id accepted")
	}
}
