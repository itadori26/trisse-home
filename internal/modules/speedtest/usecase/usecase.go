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

	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/shared/kernel"
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

// connectionTTL is how long the results of a named connection are reused.
// The background job only refreshes the module settings, so connections
// are read on demand at the same pace.
const connectionTTL = 10 * time.Minute

// Speedtest is the input port.
type Speedtest interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Results returns the latest result and the history of the last hours
	// (GetLatest, GetHistory), read through the named connection the widget
	// chose (#610) or through the module settings when connectionID is
	// empty or the connection is not usable.
	Results(ctx context.Context, connectionID string, hours int) (ResultsView, error)
	// Refresh fetches new results and checks the thresholds (job).
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port); Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Connections resolves a named connection of the integrations module to
// settings (output port, #610). Any error means the module settings are
// used instead. A connection has no thresholds.
type Connections interface {
	Settings(ctx context.Context, id string) (domain.Settings, error)
}

// SpeedtestSource reads results from Speedtest Tracker (output port),
// oldest first.
type SpeedtestSource interface {
	Results(ctx context.Context, s domain.Settings, since time.Time) ([]domain.Measurement, error)
}

type speedtest struct {
	repo   SettingsRepository
	source SpeedtestSource
	conns  Connections
	sink   kernel.ObservationSink
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *ResultsView
	// byConnection holds the results per named connection.
	byConnection map[string]ResultsView
}

// NewSpeedtest returns the use cases; conns (may be nil) resolves named
// connections, sink receives threshold observations.
func NewSpeedtest(repo SettingsRepository, source SpeedtestSource, conns Connections, sink kernel.ObservationSink, clock kernel.Clock) Speedtest {
	if sink == nil {
		sink = kernel.DiscardObservations{}
	}
	return &speedtest{repo: repo, source: source, conns: conns, sink: sink, clock: clock, byConnection: map[string]ResultsView{}}
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
		return SettingsView{}, kernel.SetupFailure("Speedtest Tracker", err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(list)
	return view(s), nil
}

// Results implements Speedtest (cache-aside).
func (uc *speedtest) Results(ctx context.Context, connectionID string, hours int) (ResultsView, error) {
	if connectionID != "" && uc.conns != nil {
		if s, err := uc.conns.Settings(ctx, connectionID); err == nil {
			return uc.window(uc.connectionResults(ctx, connectionID, s), hours), nil
		}
	}
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
	return uc.window(*cached, hours), nil
}

// window keeps the history of the last hours, at most HistoryRange.
func (uc *speedtest) window(v ResultsView, hours int) ResultsView {
	maxHours := int(HistoryRange / time.Hour)
	if hours <= 0 || hours > maxHours {
		hours = maxHours
	}
	from := uc.clock.Now().Add(-time.Duration(hours) * time.Hour)
	out := v
	out.History = nil
	for _, m := range v.History {
		if !m.At.Before(from) {
			out.History = append(out.History, m)
		}
	}
	return out
}

// connectionResults returns the cached results of a named connection or
// reads them. Like the module results, a failed read keeps the last
// results and reports the error alongside.
func (uc *speedtest) connectionResults(ctx context.Context, id string, s domain.Settings) ResultsView {
	now := uc.clock.Now()
	uc.mu.RLock()
	v, ok := uc.byConnection[id]
	uc.mu.RUnlock()
	if ok && now.Sub(v.UpdatedAt) < connectionTTL {
		return v
	}
	if list, err := uc.source.Results(ctx, s, now.Add(-HistoryRange)); err != nil {
		v.Error, v.UpdatedAt = "Speedtest Tracker is not reachable.", now
	} else {
		v = results(list, now)
	}
	uc.mu.Lock()
	uc.byConnection[id] = v
	uc.mu.Unlock()
	return v
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
	v := results(list, uc.clock.Now())
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &v
}

// results is the history with its newest measurement as the latest.
func results(list []domain.Measurement, at time.Time) ResultsView {
	v := ResultsView{History: list, UpdatedAt: at}
	if len(list) > 0 {
		latest := list[len(list)-1]
		v.Latest = &latest
	}
	return v
}
