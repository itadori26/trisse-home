package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/shared/kernel"
)

// AggregateRepository persists daily disk usage and the forecast
// settings (output port).
type AggregateRepository interface {
	Get(ctx context.Context, mount, day string) (domain.UsageAggregate, bool, error)
	Put(ctx context.Context, a domain.UsageAggregate) error
	// Since returns aggregates from day on, sorted by mount and day.
	Since(ctx context.Context, day string) ([]domain.UsageAggregate, error)
	PruneBefore(ctx context.Context, day string) error
	// ForecastSettings returns ok=false when nothing is stored.
	ForecastSettings(ctx context.Context) (domain.ForecastSettings, bool, error)
	SaveForecastSettings(ctx context.Context, s domain.ForecastSettings) error
}

// CurrentDisks reads the current usage of the file systems (output port,
// implemented by the monitor).
type CurrentDisks interface {
	Current(ctx context.Context) (SnapshotView, error)
}

// ForecastView is a forecast at the use case boundary.
type ForecastView struct {
	Status        domain.ForecastStatus
	DaysUntilFull float64
	BytesPerDay   float64
	Confidence    float64
}

// DiskForecasts is the input port of the disk forecast (#108).
type DiskForecasts interface {
	// Aggregate folds the current usage into today's aggregate, then
	// reports alerting forecasts (job).
	Aggregate(ctx context.Context) error
	// Forecasts returns the forecast per disk label.
	Forecasts(ctx context.Context) (map[string]ForecastView, error)
	Settings(ctx context.Context) (domain.ForecastSettings, error)
	SaveSettings(ctx context.Context, s domain.ForecastSettings) (domain.ForecastSettings, error)
}

type diskForecasts struct {
	disks CurrentDisks
	repo  AggregateRepository
	sink  kernel.ObservationSink // ForecastAlerter
	clock kernel.Clock
	loc   *time.Location
}

// NewDiskForecasts returns the forecast use cases; days are cut in loc.
func NewDiskForecasts(disks CurrentDisks, repo AggregateRepository, sink kernel.ObservationSink, clock kernel.Clock, loc *time.Location) DiskForecasts {
	if loc == nil {
		loc = time.Local
	}
	return &diskForecasts{disks: disks, repo: repo, sink: sink, clock: clock, loc: loc}
}

func (uc *diskForecasts) today() time.Time {
	n := uc.clock.Now().In(uc.loc)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// Aggregate implements DiskForecasts.
func (uc *diskForecasts) Aggregate(ctx context.Context) error {
	snap, err := uc.disks.Current(ctx)
	if err != nil {
		return err
	}
	today := uc.today()
	day := today.Format(domain.DayLayout)
	for _, d := range snap.Disks {
		a, _, err := uc.repo.Get(ctx, d.Label, day)
		if err != nil {
			return fmt.Errorf("load aggregate: %w", err)
		}
		a.Mount, a.Day = d.Label, day
		if err := uc.repo.Put(ctx, a.Fold(d.Usage.Used, d.Usage.Total)); err != nil {
			return fmt.Errorf("store aggregate: %w", err)
		}
	}
	if err := uc.repo.PruneBefore(ctx, today.AddDate(0, 0, -domain.AggregateRetentionDays).Format(domain.DayLayout)); err != nil {
		return fmt.Errorf("prune aggregates: %w", err)
	}
	return uc.report(ctx)
}

func (uc *diskForecasts) report(ctx context.Context) error {
	settings, err := uc.Settings(ctx)
	if err != nil {
		return err
	}
	forecasts, err := uc.compute(ctx, settings)
	if err != nil {
		return err
	}
	for label, f := range forecasts {
		o := kernel.Observation{Kind: "forecast", Key: label, Label: label, Healthy: !f.Alerting(settings), Value: f.DaysUntilFull}
		if f.Status == domain.ForecastFilling {
			o.Detail = fmt.Sprintf("At the current rate the disk %s is full in about %.0f days.", label, f.DaysUntilFull)
		}
		uc.sink.Observe(ctx, o)
	}
	return nil
}

func (uc *diskForecasts) compute(ctx context.Context, s domain.ForecastSettings) (map[string]domain.Forecast, error) {
	today := uc.today()
	aggs, err := uc.repo.Since(ctx, today.AddDate(0, 0, -s.WindowDays+1).Format(domain.DayLayout))
	if err != nil {
		return nil, fmt.Errorf("load aggregates: %w", err)
	}
	byMount := map[string][]domain.UsageAggregate{}
	for _, a := range aggs {
		byMount[a.Mount] = append(byMount[a.Mount], a)
	}
	out := make(map[string]domain.Forecast, len(byMount))
	for mount, list := range byMount {
		out[mount] = domain.ComputeForecast(list, s.WindowDays, today)
	}
	return out, nil
}

// Forecasts implements DiskForecasts.
func (uc *diskForecasts) Forecasts(ctx context.Context) (map[string]ForecastView, error) {
	settings, err := uc.Settings(ctx)
	if err != nil {
		return nil, err
	}
	forecasts, err := uc.compute(ctx, settings)
	if err != nil {
		return nil, err
	}
	out := make(map[string]ForecastView, len(forecasts))
	for label, f := range forecasts {
		out[label] = ForecastView{Status: f.Status, DaysUntilFull: f.DaysUntilFull, BytesPerDay: f.BytesPerDay, Confidence: f.Confidence}
	}
	return out, nil
}

// Settings implements DiskForecasts.
func (uc *diskForecasts) Settings(ctx context.Context) (domain.ForecastSettings, error) {
	s, ok, err := uc.repo.ForecastSettings(ctx)
	if err != nil {
		return domain.ForecastSettings{}, fmt.Errorf("load forecast settings: %w", err)
	}
	if !ok {
		return domain.DefaultForecastSettings(), nil
	}
	return s, nil
}

// SaveSettings implements DiskForecasts.
func (uc *diskForecasts) SaveSettings(ctx context.Context, s domain.ForecastSettings) (domain.ForecastSettings, error) {
	if err := s.Validate(); err != nil {
		return domain.ForecastSettings{}, err
	}
	if err := uc.repo.SaveForecastSettings(ctx, s); err != nil {
		return domain.ForecastSettings{}, fmt.Errorf("save forecast settings: %w", err)
	}
	return s, nil
}
