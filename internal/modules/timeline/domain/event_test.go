package domain

import (
	"strings"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestFromObservation(t *testing.T) {
	at := time.Unix(0, 0)
	down := FromObservation("1", at, kernel.Observation{Kind: "service", Key: "s1", Label: "NAS", Detail: "timeout"}, false)
	if down.Severity != SeverityCritical || down.Message != "NAS is offline: timeout" || down.Subject != "s1" {
		t.Fatalf("%+v", down)
	}
	up := FromObservation("2", at, kernel.Observation{Kind: "service", Key: "s1", Label: "NAS"}, true)
	if up.Severity != SeverityInfo || !strings.Contains(up.Message, "OK again") {
		t.Fatalf("%+v", up)
	}
	disk := FromObservation("3", at, kernel.Observation{Kind: "disk", Label: "/", Value: 93}, false)
	if disk.Severity != SeverityWarning || disk.Message != "Disk / is 93 % full" {
		t.Fatalf("%+v", disk)
	}
	if HealthyFor(kernel.Observation{Kind: "disk", Value: 95}) || !HealthyFor(kernel.Observation{Kind: "tls", Value: 30}) {
		t.Fatal("thresholds wrong")
	}
	if ValidRetention(0) == nil || ValidRetention(30) != nil {
		t.Fatal("retention validation wrong")
	}
}

func TestTrack(t *testing.T) {
	var tr Track
	tr, _ = tr.Advance(true)
	tr, c1 := tr.Advance(false)
	tr, c2 := tr.Advance(false)
	if c1 || !c2 {
		t.Fatal("needs two confirmations")
	}
	if _, c := tr.Advance(false); c {
		t.Fatal("no change")
	}
}
