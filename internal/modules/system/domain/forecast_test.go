package domain

import (
	"math"
	"testing"
	"time"
)

const gb = 1 << 30

// series builds daily aggregates ending today from used values in GB.
func series(today time.Time, total uint64, used ...float64) []UsageAggregate {
	out := make([]UsageAggregate, 0, len(used))
	for i, u := range used {
		day := today.AddDate(0, 0, i-len(used)+1).Format(DayLayout)
		out = append(out, UsageAggregate{Mount: "/", Day: day, Last: uint64(u * gb), Total: total})
	}
	return out
}

func TestComputeForecast(t *testing.T) {
	today := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	total := uint64(100 * gb)
	tests := []struct {
		name    string
		used    []float64
		status  ForecastStatus
		days    float64 // expected days until full (filling only)
		minConf float64
	}{
		{"linear growth", []float64{50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60}, ForecastFilling, 40, 0.99},
		{"too few days", []float64{50, 60}, ForecastInsufficient, 0, 0},
		{"flat", []float64{70, 70, 70, 70, 70}, ForecastFlat, 0, 0},
		{"shrinking", []float64{80, 78, 76, 74, 72}, ForecastFlat, 0, 0},
		{"cleanup drop", []float64{60, 62, 64, 66, 40, 41}, ForecastFlat, 0, 0},
		{"step change then growth", []float64{40, 40, 40, 60, 61, 62, 63}, ForecastFilling, 0, 0.5},
		{"almost full", []float64{95, 96, 97, 98, 99}, ForecastFilling, 1, 0.99},
		{"glacial growth", []float64{50, 50.001, 50.002, 50.003}, ForecastFlat, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := ComputeForecast(series(today, total, tc.used...), DefaultForecastWindow, today)
			if f.Status != tc.status {
				t.Fatalf("status %s, want %s (%+v)", f.Status, tc.status, f)
			}
			if tc.days > 0 && math.Abs(f.DaysUntilFull-tc.days) > 0.5 {
				t.Fatalf("days %.1f, want %.0f", f.DaysUntilFull, tc.days)
			}
			if f.Confidence < tc.minConf {
				t.Fatalf("confidence %.2f < %.2f", f.Confidence, tc.minConf)
			}
		})
	}
}

func TestForecastWindow(t *testing.T) {
	today := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	// Old growth outside a 3-day window, flat inside it.
	aggs := series(today, 100*gb, 10, 20, 30, 40, 50, 50, 50)
	if f := ComputeForecast(aggs, 3, today); f.Status != ForecastFlat || f.Days != 3 {
		t.Fatalf("%+v", f)
	}
}

func TestFoldAndSettings(t *testing.T) {
	a := UsageAggregate{}.Fold(50, 100).Fold(40, 100).Fold(45, 100)
	if a.MinUsed != 40 || a.MaxUsed != 50 || a.Last != 45 || a.Total != 100 {
		t.Fatalf("%+v", a)
	}
	if DefaultForecastSettings().Validate() != nil {
		t.Fatal("defaults invalid")
	}
	if (ForecastSettings{WindowDays: 2, AlertDays: 14}).Validate() == nil || (ForecastSettings{WindowDays: 14, AlertDays: 0}).Validate() == nil {
		t.Fatal("bounds not checked")
	}
	f := Forecast{Status: ForecastFilling, DaysUntilFull: 10}
	if !f.Alerting(DefaultForecastSettings()) || (Forecast{Status: ForecastFlat}).Alerting(DefaultForecastSettings()) {
		t.Fatal("alerting")
	}
}
