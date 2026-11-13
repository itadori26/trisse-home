// Package usecase contains the speedtest use cases.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/speedtest/domain"
	"trisse-home/internal/shared/kernel"
)

// HistoryRange is how far back results are fetched (7 days).
const HistoryRange = 7 * 24 * time.Hour

// SettingsView never contains the token, only whether one is stored.
type SettingsView struct {
	Configured  bool
	URL         string
	HasToken    bool
	MinDownMbps float64
	MinUpMbps   float64
}

// SettingsInput is the input of SaveSettings; an empty token keeps the stored one.
type SettingsInput struct {
	URL, Token             string
	MinDownMbps, MinUpMbps float64
}

// ResultsView is the cached state: latest result and history (oldest first).
type ResultsView struct {
	Latest    *domain.Measurement
	History   []domain.Measurement
	UpdatedAt time.Time
	Error     string
}

// Speedtest is the input port.
type Speedtest interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Results returns the latest result and the history of the last hours
	// (GetLatest, GetHistory).
	Results(ctx context.Context, hours int) (ResultsView, error)
	// Refresh fetches new results and checks the thresholds (job).
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port); Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// SpeedtestSource reads results from Speedtest Tracker (output port),
// oldest first.
type SpeedtestSource interface {
	Results(ctx context.Context, s domain.Settings, since time.Time) ([]domain.Measurement, error)
}

type speedtest struct {
	repo   SettingsRepository
	source SpeedtestSource
	sink   kernel.ObservationSink
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *ResultsView
}

// NewSpeedtest returns the use cases; sink receives threshold observations.
func NewSpeedtest(repo SettingsRepository, source SpeedtestSource, sink kernel.ObservationSink, clock kernel.Clock) Speedtest {
	if sink == nil {
		sink = kernel.DiscardObservations{}
	}
	return &speedtest{repo: repo, source: source, sink: sink, clock: clock}
}

// Settings implements Speedtest.
func (uc *speedtest) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return view(s), nil
}

func view(s domain.Settings) SettingsView {
	return SettingsView{Configured: true, URL: s.URL, HasToken: s.Token != "", MinDownMbps: s.MinDownMbps, MinUpMbps: s.MinUpMbps}
}

// SaveSettings implements Speedtest; the connection is tested first.
func (uc *speedtest) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	token := in.Token
	if token == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			if kernel.TargetChanged([]string{old.URL}, []string{in.URL}) {
				return SettingsView{}, kernel.SecretRequired("token")
			}
			token = old.Token
		}
	}
	s, err := domain.NewSettings(in.URL, token, in.MinDownMbps, in.MinUpMbps)
	if err != nil {
		return SettingsView{}, err
	}
	list, err := uc.source.Results(ctx, s, uc.clock.Now().Add(-HistoryRange))
	if err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(list)
	return view(s), nil
}

// Results implements Speedtest (cache-aside).
func (uc *speedtest) Results(ctx context.Context, hours int) (ResultsView, error) {
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached == nil {
		if _, err := uc.repo.Load(ctx); err != nil {
			return ResultsView{}, err
		}
		_ = uc.Refresh(ctx) // reported through ResultsView.Error
		uc.mu.RLock()
		cached = uc.cached
		uc.mu.RUnlock()
		if cached == nil {
			return ResultsView{}, kernel.NewPublicError(kernel.ErrUnavailable, "Speedtest Tracker is not reachable.")
		}
	}
	maxHours := int(HistoryRange / time.Hour)
	if hours <= 0 || hours > maxHours {
		hours = maxHours
	}
	from := uc.clock.Now().Add(-time.Duration(hours) * time.Hour)
	out := *cached
	out.History = nil
	for _, m := range cached.History {
		if !m.At.Before(from) {
			out.History = append(out.History, m)
		}
	}
	return out, nil
}

// Refresh implements Speedtest.
func (uc *speedtest) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	list, err := uc.source.Results(ctx, s, uc.clock.Now().Add(-HistoryRange))
	if err != nil {
		uc.mu.Lock()
		v := ResultsView{}
		if uc.cached != nil {
			v = *uc.cached
		}
		v.Error = "Speedtest Tracker is not reachable."
		uc.cached = &v
		uc.mu.Unlock()
		return fmt.Errorf("query speedtest tracker: %w", err)
	}
	uc.store(list)
	if s.HasThresholds() && len(list) > 0 {
		latest := list[len(list)-1]
		healthy, detail := s.Judge(latest)
		uc.sink.Observe(ctx, kernel.Observation{Kind: "speedtest", Key: "internet", Label: "Internet", Healthy: healthy,
			Value: domain.Mbps(latest.DownloadBps), Detail: detail})
	}
	return nil
}

func (uc *speedtest) store(list []domain.Measurement) {
	v := ResultsView{History: list, UpdatedAt: uc.clock.Now()}
	if len(list) > 0 {
		latest := list[len(list)-1]
		v.Latest = &latest
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &v
}
