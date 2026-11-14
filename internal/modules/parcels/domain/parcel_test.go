package domain

import (
	"testing"
	"time"
)

func TestParcel(t *testing.T) {
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	p, err := NewParcel("1", " 0034 0434 1234 5678 9012 ", CarrierDHL, " Shoes ", now)
	if err != nil || p.TrackingNumber != "00340434123456789012" || p.Label != "Shoes" {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := NewParcel("1", "abc", CarrierDHL, "", now); err == nil {
		t.Fatal("short number accepted")
	}
	if _, err := NewParcel("1", "JJD000390007912345", "fedex", "", now); err == nil {
		t.Fatal("unsupported carrier accepted")
	}
	p.Status = Status{Code: StatusDelivered, At: now.Add(-4 * 24 * time.Hour)}
	if p.Visible(now, 3) || !p.Visible(now, 5) {
		t.Fatal("visibility wrong")
	}
	ps := []Parcel{p, {ID: "2", CreatedAt: now}, {ID: "3", CreatedAt: now.Add(-time.Hour)}}
	SortForDisplay(ps)
	if ps[0].ID != "3" || ps[2].ID != "1" {
		t.Fatalf("order %v %v %v", ps[0].ID, ps[1].ID, ps[2].ID)
	}
	if _, err := NewSettings("k", 31); err == nil {
		t.Fatal("31 days accepted")
	}
}
