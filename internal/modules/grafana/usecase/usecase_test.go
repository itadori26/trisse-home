package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/grafana/domain"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}

func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeClient struct {
	alerts    []domain.Alert
	alertCall int
	searchErr error
	data      domain.PanelData
	dataErr   error
}

func (f *fakeClient) PanelData(context.Context, domain.Settings, domain.DataRequest) (domain.PanelData, error) {
	return f.data, f.dataErr
}

func (f *fakeClient) Alerts(context.Context, domain.Settings) ([]domain.Alert, error) {
	f.alertCall++
	return f.alerts, nil
}

func (f *fakeClient) Search(context.Context, domain.Settings, string) ([]domain.Dashboard, error) {
	return []domain.Dashboard{{UID: "node"}}, f.searchErr
}

func (f *fakeClient) Panels(context.Context, domain.Settings, string) ([]domain.Panel, error) {
	return []domain.Panel{{ID: 1}}, nil
}

func (f *fakeClient) Render(context.Context, domain.Settings, domain.RenderRequest) (domain.Image, error) {
	return domain.Image{ContentType: "image/png"}, nil
}

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func TestSettingsKeepTheTokenAndTestTheConnection(t *testing.T) {
	repo, client := &memRepo{}, &fakeClient{}
	uc := New(repo, client, nil, &clock{now: time.Now()})
	if v, _ := uc.Settings(context.Background()); v.Configured {
		t.Fatal("not configured yet")
	}
	if _, err := uc.SaveSettings(context.Background(), SettingsInput{URL: "http://g:3000", Token: "glsa_1"}); err != nil {
		t.Fatal(err)
	}
	v, err := uc.SaveSettings(context.Background(), SettingsInput{URL: "http://grafana:3000"})
	if err != nil || !v.HasToken || repo.s.Token != "glsa_1" || repo.s.URL != "http://grafana:3000" {
		t.Fatalf("view = %+v, stored = %+v, %v", v, repo.s, err)
	}
	client.searchErr = errors.New("the token was rejected")
	if _, err := uc.SaveSettings(context.Background(), SettingsInput{URL: "http://other", Token: "bad"}); err == nil || repo.s.Token != "glsa_1" {
		t.Fatalf("a failing connection must not be saved: %v", err)
	}
}

func TestAlertsAreSortedAndCached(t *testing.T) {
	now := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	c := &clock{now: now}
	client := &fakeClient{alerts: []domain.Alert{
		{Name: "pending", State: domain.Pending, Since: now.Add(-time.Minute)},
		{Name: "old", State: domain.Firing, Since: now.Add(-time.Hour)},
		{Name: "new", State: domain.Firing, Since: now.Add(-2 * time.Minute)},
	}}
	uc := New(&memRepo{s: &domain.Settings{URL: "http://g", Token: "t"}}, client, nil, c)
	list, err := uc.Alerts(context.Background(), "")
	if err != nil || list[0].Name != "new" || list[1].Name != "old" || list[2].Name != "pending" {
		t.Fatalf("alerts = %+v, %v", list, err)
	}
	_, _ = uc.Alerts(context.Background(), "")
	c.now = now.Add(time.Minute)
	_, _ = uc.Alerts(context.Background(), "")
	if client.alertCall != 2 {
		t.Fatalf("Grafana asked %d times, want 2", client.alertCall)
	}
}

func TestNotConfiguredAndBadUID(t *testing.T) {
	uc := New(&memRepo{}, &fakeClient{}, nil, &clock{now: time.Now()})
	if _, err := uc.Alerts(context.Background(), ""); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
	if _, err := uc.Panels(context.Background(), "", "a/b"); !errors.Is(err, domain.ErrBadDashboard) {
		t.Fatalf("err = %v", err)
	}
	if uc.URL(context.Background(), "") != "" {
		t.Fatal("no URL without settings")
	}
}

func TestPanelDataKeepsItsOwnErrors(t *testing.T) {
	repo := &memRepo{s: &domain.Settings{URL: "http://g", Token: "t"}}
	req := domain.DataRequest{UID: "node", PanelID: 1, Range: "6h"}
	for name, c := range map[string]struct {
		err, want error
	}{
		"missing panel": {domain.ErrBadDashboard, domain.ErrBadDashboard},
		"no values":     {domain.ErrNoData, domain.ErrNoData},
		"other":         {errors.New("boom"), domain.ErrUnreachable},
	} {
		uc := New(repo, &fakeClient{dataErr: c.err}, nil, &clock{now: time.Now()})
		if _, err := uc.PanelData(context.Background(), "", req); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	want := domain.PanelData{Title: "CPU", Series: []domain.Series{{Name: "cpu", Points: []domain.Point{{Value: 1}}}}}
	uc := New(repo, &fakeClient{data: want}, nil, &clock{now: time.Now()})
	if got, err := uc.PanelData(context.Background(), "", req); err != nil || got.Title != "CPU" {
		t.Fatalf("got %+v %v", got, err)
	}
	if _, err := New(&memRepo{}, &fakeClient{}, nil, &clock{}).PanelData(context.Background(), "", req); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
}

// conns resolves "office" to another Grafana; everything else fails, like
// a missing copy or a connection of another service.
type conns struct{}

func (conns) Settings(_ context.Context, id string) (domain.Settings, error) {
	if id == "office" {
		return domain.Settings{URL: "http://office", Token: "t"}, nil
	}
	return domain.Settings{}, errors.New("no usable connection")
}

// recordingClient remembers which Grafana was asked for alerts.
type recordingClient struct {
	fakeClient
	urls []string
}

func (r *recordingClient) Alerts(ctx context.Context, s domain.Settings) ([]domain.Alert, error) {
	r.urls = append(r.urls, s.URL)
	return r.fakeClient.Alerts(ctx, s)
}

func TestAlertsThroughConnection(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{s: &domain.Settings{URL: "http://home", Token: "t"}}
	client := &recordingClient{}
	uc := New(repo, client, conns{}, &clock{now: time.Now()})
	for _, id := range []string{"office", "", "missing"} {
		if _, err := uc.Alerts(ctx, id); err != nil {
			t.Fatalf("%q: %v", id, err)
		}
	}
	// "missing" falls back to the module settings, cached on its own key.
	if want := "http://office http://home http://home"; strings.Join(client.urls, " ") != want {
		t.Fatalf("asked %v, want %s", client.urls, want)
	}
	_, _ = uc.Alerts(ctx, "office")
	if len(client.urls) != 3 {
		t.Fatal("each connection is cached")
	}
	if got := uc.URL(ctx, "office"); got != "http://office" {
		t.Fatalf("links point to the connection: %q", got)
	}
	repo.s = nil
	if _, err := uc.Dashboards(ctx, "office", ""); err != nil {
		t.Fatalf("a connection works without module settings: %v", err)
	}
	if _, err := uc.Dashboards(ctx, "missing", ""); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("no connection and no settings: %v", err)
	}
}
