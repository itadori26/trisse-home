package domain

import (
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestEvaluate(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	fresh := now.Add(-10 * time.Minute)
	tests := []struct {
		name string
		r    Report
		want Status
	}{
		{"healthy ATA", Report{Protocol: ProtocolATA, At: fresh, Passed: ptr(true), Temperature: ptr(35),
			Attributes: []Attribute{{ID: 5, Name: "Reallocated_Sector_Ct", Raw: 0}}}, StatusOK},
		{"reallocated sectors", Report{Protocol: ProtocolATA, At: fresh, Passed: ptr(true),
			Attributes: []Attribute{{ID: 5, Raw: 8}}}, StatusWarning},
		{"pending sectors", Report{Protocol: ProtocolATA, At: fresh, Passed: ptr(true),
			Attributes: []Attribute{{ID: 197, Raw: 1}}}, StatusWarning},
		{"failing attribute", Report{Protocol: ProtocolATA, At: fresh, Passed: ptr(true),
			Attributes: []Attribute{{ID: 1, Name: "Raw_Read_Error_Rate", FailingNow: true}}}, StatusCritical},
		{"self-assessment failed", Report{Protocol: ProtocolATA, At: fresh, Passed: ptr(false)}, StatusCritical},
		{"hot HDD", Report{Protocol: ProtocolATA, At: fresh, Passed: ptr(true), Temperature: ptr(62)}, StatusWarning},
		{"warm NVMe is fine", Report{Protocol: ProtocolNVMe, At: fresh, Passed: ptr(true), Temperature: ptr(62), NVMe: &NVMeHealth{AvailableSpare: 100, SpareThreshold: 10}}, StatusOK},
		{"NVMe critical warning", Report{Protocol: ProtocolNVMe, At: fresh, Passed: ptr(true), NVMe: &NVMeHealth{CriticalWarning: 4, AvailableSpare: 100}}, StatusCritical},
		{"NVMe spare low", Report{Protocol: ProtocolNVMe, At: fresh, Passed: ptr(true), NVMe: &NVMeHealth{AvailableSpare: 5, SpareThreshold: 10}}, StatusCritical},
		{"NVMe worn", Report{Protocol: ProtocolNVMe, At: fresh, Passed: ptr(true), NVMe: &NVMeHealth{AvailableSpare: 100, SpareThreshold: 10, PercentageUsed: 93}}, StatusWarning},
		{"NVMe media errors", Report{Protocol: ProtocolNVMe, At: fresh, Passed: ptr(true), NVMe: &NVMeHealth{AvailableSpare: 100, MediaErrors: 2}}, StatusWarning},
		{"no data", Report{Protocol: ProtocolOther, At: fresh}, StatusUnknown},
		{"stale but failing", Report{Protocol: ProtocolATA, At: now.Add(-3 * time.Hour), Passed: ptr(false)}, StatusUnknown},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := Evaluate(tc.r, now, DefaultStaleAfter)
			if d.Status != tc.want {
				t.Fatalf("status %s, want %s (%+v)", d.Status, tc.want, d.Findings)
			}
			if tc.want != StatusOK && len(d.Findings) == 0 {
				t.Fatal("findings missing")
			}
		})
	}
}

func TestObservation(t *testing.T) {
	now := time.Now()
	d := Evaluate(Report{Device: "/dev/sda", Model: "WD Red", At: now, Passed: ptr(true), Attributes: []Attribute{{ID: 5, Raw: 3}}}, now, DefaultStaleAfter)
	o, ok := d.Observation()
	if !ok || o.Healthy || o.Kind != "diskhealth" || o.Label != "WD Red (/dev/sda)" || o.Detail != "3 reallocated sectors" {
		t.Fatalf("%+v %v", o, ok)
	}
	if *d.Reallocated != 3 {
		t.Fatal("reallocated count")
	}
	unknown := Evaluate(Report{Device: "/dev/sdb", At: now}, now, DefaultStaleAfter)
	if _, ok := unknown.Observation(); ok {
		t.Fatal("unknown disks must not be observed")
	}
}
