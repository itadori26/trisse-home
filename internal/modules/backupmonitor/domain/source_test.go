package domain

import (
	"testing"
	"time"
)

func TestSourceStates(t *testing.T) {
	t0 := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	s, err := NewSource("1", "NAS", 24*time.Hour, HashToken("x"), t0)
	if err != nil {
		t.Fatal(err)
	}
	if s.StateAt(t0.Add(time.Hour)) != StateWaiting || s.StateAt(t0.Add(25*time.Hour)) != StateOverdue {
		t.Fatal("waiting/overdue before first run wrong")
	}
	s.RecordRun(true, " ok ", t0.Add(2*time.Hour))
	if s.StateAt(t0.Add(3*time.Hour)) != StateOK || s.LastMessage != "ok" {
		t.Fatal("ok after success wrong")
	}
	if s.StateAt(t0.Add(27*time.Hour)) != StateOverdue {
		t.Fatal("overdue after max age wrong")
	}
	s.RecordRun(false, "disk full", t0.Add(4*time.Hour))
	if s.StateAt(t0.Add(5*time.Hour)) != StateFailed {
		t.Fatal("failure must win over older success")
	}
	s.RecordRun(true, "", t0.Add(6*time.Hour))
	if s.StateAt(t0.Add(7*time.Hour)) != StateOK {
		t.Fatal("new success must clear failure")
	}
}

func TestSourceValidation(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name   string
		maxAge time.Duration
	}{{"", time.Hour}, {"x", time.Minute}, {"x", 61 * 24 * time.Hour}} {
		if _, err := NewSource("1", tc.name, tc.maxAge, "h", now); err == nil {
			t.Errorf("expected error for %+v", tc)
		}
	}
}
