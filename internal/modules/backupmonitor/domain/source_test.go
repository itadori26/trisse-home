package domain

import (
	"testing"
	"time"
)

func TestSourceStates(t *testing.T) {
	t0 := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	s, err := NewSource("1", "NAS", Schedule{MaxAge: 24 * time.Hour}, HashToken("x"), t0)
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
		if _, err := NewSource("1", tc.name, Schedule{MaxAge: tc.maxAge}, "h", now); err == nil {
			t.Errorf("expected error for %+v", tc)
		}
	}
}

func TestHeartbeatSchedules(t *testing.T) {
	t0 := time.Date(2026, 11, 11, 0, 30, 0, 0, time.UTC)
	daily, err := NewSource("1", "Nightly job", Schedule{Kind: "check", Cron: "0 3 * * *", Grace: 30 * time.Minute}, HashToken("x"), t0)
	if err != nil {
		t.Fatal(err)
	}
	if daily.StateAt(t0.Add(3*time.Hour)) != StateWaiting { // 03:30, within grace
		t.Fatal("must wait within grace")
	}
	if daily.StateAt(t0.Add(3*time.Hour+time.Minute)) != StateOverdue { // 03:31
		t.Fatal("must be overdue after grace")
	}
	daily.RecordStart(t0.Add(2*time.Hour + 30*time.Minute))
	if daily.StateAt(t0.Add(2*time.Hour+40*time.Minute)) != StateRunning {
		t.Fatal("must be running after start")
	}
	daily.RecordRun(true, "", t0.Add(2*time.Hour+50*time.Minute))
	if daily.LastDuration != 20*time.Minute || daily.StateAt(t0.Add(10*time.Hour)) != StateOK {
		t.Fatalf("duration %v state %s", daily.LastDuration, daily.StateAt(t0.Add(10*time.Hour)))
	}
	if daily.StateAt(t0.Add(27*time.Hour+time.Minute)) != StateOverdue { // next night missed
		t.Fatal("next run missed must be overdue")
	}
	interval, _ := NewSource("2", "Backup", Schedule{MaxAge: 24 * time.Hour, Grace: time.Hour}, HashToken("y"), t0)
	if interval.Kind != "backup" || interval.StateAt(t0.Add(24*time.Hour+30*time.Minute)) != StateWaiting || interval.StateAt(t0.Add(25*time.Hour+time.Minute)) != StateOverdue {
		t.Fatal("interval with grace wrong")
	}
	for _, sch := range []Schedule{{Kind: "x", MaxAge: time.Hour}, {Cron: "bad"}, {MaxAge: time.Minute}, {MaxAge: time.Hour, Grace: -time.Second}} {
		if _, err := NewSource("3", "x", sch, HashToken("z"), t0); err == nil {
			t.Errorf("accepted %+v", sch)
		}
	}
}
