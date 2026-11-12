package domain

import (
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestTrackDebounce(t *testing.T) {
	var tr Track
	var notified bool
	tr, notified = tr.Advance(true)
	if notified || !tr.Healthy {
		t.Fatal("first observation must only set the baseline")
	}
	tr, notified = tr.Advance(false)
	if notified {
		t.Fatal("single failure must not notify")
	}
	tr, notified = tr.Advance(true)
	if notified {
		t.Fatal("flapping must not notify")
	}
	tr, _ = tr.Advance(false)
	tr, notified = tr.Advance(false)
	if !notified || tr.Healthy {
		t.Fatal("two consecutive failures must notify")
	}
	tr, notified = tr.Advance(false)
	if notified {
		t.Fatal("unchanged state must not notify again")
	}
	tr, _ = tr.Advance(true)
	if _, notified = tr.Advance(true); !notified {
		t.Fatal("recovery must notify")
	}
}

func TestRulesJudge(t *testing.T) {
	r := DefaultRules()
	cases := []struct {
		o        kernel.Observation
		relevant bool
		healthy  bool
	}{
		{kernel.Observation{Kind: "disk", Value: 95}, true, false},
		{kernel.Observation{Kind: "disk", Value: 50}, true, true},
		{kernel.Observation{Kind: "tls", Value: 10}, true, false},
		{kernel.Observation{Kind: "tls", Value: 30}, true, true},
		{kernel.Observation{Kind: "service", Healthy: false}, true, false},
		{kernel.Observation{Kind: "diskhealth", Healthy: false}, true, false},
		{kernel.Observation{Kind: "forecast", Healthy: false, Value: 9}, true, false},
		{kernel.Observation{Kind: "speedtest", Healthy: false}, true, false},
		{kernel.Observation{Kind: "image", Healthy: false}, true, false},
		{kernel.Observation{Kind: "other"}, false, true},
	}
	for _, c := range cases {
		rel, h := r.Judge(c.o)
		if rel != c.relevant || h != c.healthy {
			t.Errorf("%+v: got %v %v", c.o, rel, h)
		}
	}
	r.Services = false
	if rel, _ := r.Judge(kernel.Observation{Kind: "service"}); rel {
		t.Error("disabled rule still relevant")
	}
}

func TestSettingsValidation(t *testing.T) {
	if _, err := NewSettings(true, "", "", "", DefaultRules()); err == nil {
		t.Fatal("enabled without topic accepted")
	}
	s, err := NewSettings(true, " https://ntfy.example.com/ ", "home-alerts", "", DefaultRules())
	if err != nil || s.ServerURL != "https://ntfy.example.com" {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	bad := DefaultRules()
	bad.DiskThreshold = 10
	if _, err := NewSettings(false, "", "", "", bad); err == nil {
		t.Fatal("threshold out of range accepted")
	}
	if m := MessageFor(kernel.Observation{Kind: "disk", Label: "/", Value: 93.6}, false); m.Title != "Disk / is 94 % full" || m.Priority != PriorityHigh {
		t.Fatalf("message %+v", m)
	}
}
