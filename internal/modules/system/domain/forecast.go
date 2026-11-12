package domain

import (
	"math"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Disk forecast limits (#108).
const (
	AggregateRetentionDays = 90
	DefaultForecastWindow  = 14
	DefaultForecastAlert   = 14
	minForecastDays        = 3
	maxForecastWindow      = 90
	maxForecastAlert       = 365
	// Forecasts beyond ten years are treated as flat.
	maxDaysUntilFull = 3650
)

// DayLayout formats aggregate days.
const DayLayout = "2006-01-02"

// UsageAggregate is the daily usage of one file system.
type UsageAggregate struct {
	Mount   string
	Day     string // DayLayout, local time
	MinUsed uint64
	MaxUsed uint64
	Last    uint64
	Total   uint64
}

// Fold adds a reading to the aggregate of its day.
func (a UsageAggregate) Fold(used, total uint64) UsageAggregate {
	if a.MaxUsed == 0 && a.MinUsed == 0 && a.Last == 0 {
		a.MinUsed, a.MaxUsed = used, used
	}
	a.MinUsed = min(a.MinUsed, used)
	a.MaxUsed = max(a.MaxUsed, used)
	a.Last, a.Total = used, total
	return a
}

// ForecastSettings configures the regression window and the alert.
type ForecastSettings struct {
	WindowDays int
	AlertDays  int
}

// DefaultForecastSettings returns the defaults.
func DefaultForecastSettings() ForecastSettings {
	return ForecastSettings{WindowDays: DefaultForecastWindow, AlertDays: DefaultForecastAlert}
}

// Validate checks the settings.
func (s ForecastSettings) Validate() error {
	if s.WindowDays < minForecastDays || s.WindowDays > maxForecastWindow {
		return kernel.NewValidationError("windowDays", "The forecast window must be between 3 and 90 days.")
	}
	if s.AlertDays < 1 || s.AlertDays > maxForecastAlert {
		return kernel.NewValidationError("alertDays", "The forecast warning must be between 1 and 365 days.")
	}
	return nil
}

// ForecastStatus says whether a forecast exists.
type ForecastStatus string

// Forecast statuses.
const (
	// ForecastFilling: usage grows, DaysUntilFull is set.
	ForecastFilling ForecastStatus = "filling"
	// ForecastFlat: usage is flat or shrinking; no forecast.
	ForecastFlat ForecastStatus = "flat"
	// ForecastInsufficient: fewer than three days of data.
	ForecastInsufficient ForecastStatus = "insufficient"
)

// Forecast predicts when a file system is full.
type Forecast struct {
	Status        ForecastStatus
	DaysUntilFull float64
	BytesPerDay   float64
	// Confidence is the R² of the regression (0..1).
	Confidence float64
	Days       int // days of data used
}

// ComputeForecast fits a line through the last usage of each day within the
// window (aggregates sorted by day, all of the same mount).
func ComputeForecast(aggs []UsageAggregate, windowDays int, today time.Time) Forecast {
	from := today.AddDate(0, 0, -windowDays+1).Format(DayLayout)
	var xs, ys []float64
	var last UsageAggregate
	for _, a := range aggs {
		if a.Day < from {
			continue
		}
		d, err := time.Parse(DayLayout, a.Day)
		if err != nil {
			continue
		}
		xs = append(xs, d.Sub(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)).Hours()/24)
		ys = append(ys, float64(a.Last))
		last = a
	}
	f := Forecast{Status: ForecastInsufficient, Days: len(xs)}
	if len(xs) < minForecastDays {
		return f
	}
	slope, _, r2 := regression(xs, ys)
	f.BytesPerDay, f.Confidence = slope, r2
	// Less than 0.01 % of the size per day is noise, not growth.
	free := float64(last.Total) - float64(last.Last)
	if slope <= float64(last.Total)*0.0001 || last.Total == 0 {
		f.Status = ForecastFlat
		return f
	}
	days := free / slope
	if days > maxDaysUntilFull {
		f.Status = ForecastFlat
		return f
	}
	f.Status, f.DaysUntilFull = ForecastFilling, math.Max(0, days)
	return f
}

// regression returns slope, intercept and R² of a least-squares fit.
func regression(xs, ys []float64) (slope, intercept, r2 float64) {
	n := float64(len(xs))
	var sx, sy, sxx, sxy float64
	for i := range xs {
		sx += xs[i]
		sy += ys[i]
		sxx += xs[i] * xs[i]
		sxy += xs[i] * ys[i]
	}
	den := n*sxx - sx*sx
	if den == 0 {
		return 0, sy / n, 0
	}
	slope = (n*sxy - sx*sy) / den
	intercept = (sy - slope*sx) / n
	mean := sy / n
	var ssTot, ssRes float64
	for i := range xs {
		p := slope*xs[i] + intercept
		ssRes += (ys[i] - p) * (ys[i] - p)
		ssTot += (ys[i] - mean) * (ys[i] - mean)
	}
	if ssTot == 0 {
		return slope, intercept, 1
	}
	return slope, intercept, 1 - ssRes/ssTot
}

// Alerting reports whether the forecast is below the warning threshold.
func (f Forecast) Alerting(s ForecastSettings) bool {
	return f.Status == ForecastFilling && f.DaysUntilFull < float64(s.AlertDays)
}
