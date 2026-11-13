package domain

import (
	"math"
	"testing"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		lat, lon float64
		wantErr  bool
		wantKey  string
	}{
		{53.1435, 8.2146, false, "53.14,8.21"},
		{-91, 0, true, ""},
		{0, 181, true, ""},
		{math.NaN(), 0, true, ""},
	}
	for _, tt := range tests {
		loc, err := NewLocation(tt.lat, tt.lon)
		if (err != nil) != tt.wantErr || (err == nil && loc.Key() != tt.wantKey) {
			t.Errorf("NewLocation(%v,%v) = %v, %v", tt.lat, tt.lon, loc.Key(), err)
		}
	}
}

func TestNormalizeDegrees(t *testing.T) {
	for in, want := range map[float64]float64{0: 0, 90: 90, 359.5: 359.5, 360: 0, 450: 90, -45: 315, -360: 0, 720.5: 0.5} {
		if got := NormalizeDegrees(in); got != want {
			t.Errorf("NormalizeDegrees(%v) = %v, want %v", in, got, want)
		}
	}
}
