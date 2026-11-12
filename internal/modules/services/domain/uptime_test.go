package domain

import (
	"testing"
	"time"
)

func TestUptime(t *testing.T) {
	if _, ok := Uptime(nil); ok {
		t.Fatal("no data must not be 0 %")
	}
	aggs := []DailyAggregate{{Checks: 1440, Online: 1440}, {Checks: 1440, Online: 1380, Maintenance: 60}}
	if u, ok := Uptime(aggs); !ok || u != 100 {
		t.Fatalf("maintenance must not count: %v", u)
	}
	if u, _ := Uptime([]DailyAggregate{{Checks: 100, Online: 99}}); u != 99 {
		t.Fatalf("got %v", u)
	}
}

func TestIncidentTrackDebounces(t *testing.T) {
	t0 := time.Unix(0, 0)
	var tr IncidentTrack
	step := func(online bool, minute int, cause string) Transition {
		var tn Transition
		tr, tn = tr.Advance(online, t0.Add(time.Duration(minute)*time.Minute), cause)
		return tn
	}
	if step(true, 0, "").Changed {
		t.Fatal("baseline must not change")
	}
	if step(false, 1, "timeout").Changed || step(true, 2, "").Changed {
		t.Fatal("a single failure is flapping")
	}
	step(false, 3, "timeout")
	tn := step(false, 4, "refused")
	if !tn.Changed || tn.Online || !tn.At.Equal(t0.Add(3*time.Minute)) || tn.Cause != "timeout" {
		t.Fatalf("outage: %+v", tn)
	}
	step(true, 10, "")
	if tn := step(true, 11, ""); !tn.Changed || !tn.Online || !tn.At.Equal(t0.Add(10*time.Minute)) {
		t.Fatalf("recovery: %+v", tn)
	}
	in := Incident{Start: t0, End: t0.Add(time.Hour)}
	if in.Open() || in.Duration(t0) != time.Hour {
		t.Fatal("incident duration wrong")
	}
}
