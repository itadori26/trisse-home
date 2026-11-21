package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"waylight/internal/modules/grafana/domain"
	"waylight/internal/shared/kernel"
)

// alertsTTL bounds how often Grafana is asked for alerts.
const alertsTTL = 30 * time.Second

// SettingsView never contains the token.
type SettingsView struct {
	Configured bool
	URL        string
	HasToken   bool
}

// SettingsInput is the input of SaveSettings; an empty token keeps the stored one.
type SettingsInput struct {
	URL, Token string
}

// Grafana is the input port of the module (#574). The read methods take
// the named connection the widget chose (#610); an empty connectionID, or
// a connection that is not usable, reads through the module settings.
type Grafana interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Alerts are the firing and pending alerts, firing first, then newest.
	Alerts(ctx context.Context, connectionID string) ([]domain.Alert, error)
	Dashboards(ctx context.Context, connectionID, query string) ([]domain.Dashboard, error)
	Panels(ctx context.Context, connectionID, uid string) ([]domain.Panel, error)
	Render(ctx context.Context, connectionID string, r domain.RenderRequest) (domain.Image, error)
	// PanelData are the panel's values without the image renderer (#579).
	PanelData(ctx context.Context, connectionID string, r domain.DataRequest) (domain.PanelData, error)
	// URL is the address of Grafana, for links; empty while not configured.
	URL(ctx context.Context, connectionID string) string
}

// Connections resolves a named connection of the integrations module to
// settings (output port, #610). Any error means the module settings are
// used instead.
type Connections interface {
	Settings(ctx context.Context, id string) (domain.Settings, error)
}

// SettingsRepository persists the settings (output port). Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Client talks to the Grafana HTTP API (output port).
type Client interface {
	Alerts(ctx context.Context, s domain.Settings) ([]domain.Alert, error)
	Search(ctx context.Context, s domain.Settings, query string) ([]domain.Dashboard, error)
	Panels(ctx context.Context, s domain.Settings, uid string) ([]domain.Panel, error)
	Render(ctx context.Context, s domain.Settings, r domain.RenderRequest) (domain.Image, error)
	PanelData(ctx context.Context, s domain.Settings, r domain.DataRequest) (domain.PanelData, error)
}

type cachedAlerts struct {
	alerts  []domain.Alert
	fetched time.Time
}

type grafana struct {
	repo   SettingsRepository
	client Client
	conns  Connections
	clock  kernel.Clock
	mu     sync.Mutex
	// alerts holds one list per connection ID; "" is the module settings.
	alerts map[string]cachedAlerts
}

// New returns the use cases; conns may be nil, then only the module
// settings are used.
func New(repo SettingsRepository, client Client, conns Connections, clock kernel.Clock) Grafana {
	return &grafana{repo: repo, client: client, conns: conns, clock: clock, alerts: map[string]cachedAlerts{}}
}

// load returns the settings of the named connection, falling back to the
// module settings while the connection does not exist or is not usable.
func (uc *grafana) load(ctx context.Context, connectionID string) (domain.Settings, error) {
	if connectionID != "" && uc.conns != nil {
		if s, err := uc.conns.Settings(ctx, connectionID); err == nil {
			return s, nil
		}
	}
	return uc.repo.Load(ctx)
}

// Settings implements Grafana.
func (uc *grafana) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, HasToken: true}, nil
}

// SaveSettings implements Grafana; the token is tested with a dashboard
// search, which a Viewer service account may do.
func (uc *grafana) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	token := in.Token
	if token == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			token = old.Token
		}
	}
	s, err := domain.NewSettings(in.URL, token)
	if err != nil {
		return SettingsView{}, err
	}
	if _, err := uc.client.Search(ctx, s, ""); err != nil {
		return SettingsView{}, kernel.SetupFailure("Grafana", err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	clear(uc.alerts)
	uc.mu.Unlock()
	return SettingsView{Configured: true, URL: s.URL, HasToken: true}, nil
}

// Alerts implements Grafana.
func (uc *grafana) Alerts(ctx context.Context, connectionID string) ([]domain.Alert, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.clock.Now()
	if c, ok := uc.alerts[connectionID]; ok && now.Sub(c.fetched) < alertsTTL {
		return c.alerts, nil
	}
	list, err := uc.client.Alerts(ctx, s)
	if err != nil {
		return nil, domain.ErrUnreachable
	}
	slices.SortStableFunc(list, func(a, b domain.Alert) int {
		if a.State != b.State {
			if a.State == domain.Firing {
				return -1
			}
			return 1
		}
		return b.Since.Compare(a.Since)
	})
	if list == nil {
		list = []domain.Alert{}
	}
	uc.alerts[connectionID] = cachedAlerts{alerts: list, fetched: now}
	return list, nil
}

// Dashboards implements Grafana.
func (uc *grafana) Dashboards(ctx context.Context, connectionID, query string) ([]domain.Dashboard, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	list, err := uc.client.Search(ctx, s, query)
	if err != nil {
		return nil, domain.ErrUnreachable
	}
	return list, nil
}

// Panels implements Grafana.
func (uc *grafana) Panels(ctx context.Context, connectionID, uid string) ([]domain.Panel, error) {
	if !domain.ValidUID(uid) {
		return nil, domain.ErrBadDashboard
	}
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	return uc.client.Panels(ctx, s, uid)
}

// Render implements Grafana.
func (uc *grafana) Render(ctx context.Context, connectionID string, r domain.RenderRequest) (domain.Image, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return domain.Image{}, err
	}
	return uc.client.Render(ctx, s, r)
}

// PanelData implements Grafana. A missing dashboard or panel and a panel
// without readable values keep their own error; anything else means Grafana
// did not answer properly.
func (uc *grafana) PanelData(ctx context.Context, connectionID string, r domain.DataRequest) (domain.PanelData, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return domain.PanelData{}, err
	}
	d, err := uc.client.PanelData(ctx, s, r)
	switch {
	case errors.Is(err, domain.ErrBadDashboard), errors.Is(err, domain.ErrNoData):
		return domain.PanelData{}, err
	case err != nil:
		return domain.PanelData{}, domain.ErrUnreachable
	}
	return d, nil
}

// URL implements Grafana.
func (uc *grafana) URL(ctx context.Context, connectionID string) string {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return ""
	}
	return s.URL
}
