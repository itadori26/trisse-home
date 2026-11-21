package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeSource struct {
	list  []domain.Measurement
	err   error
	token string
}

func (f *fakeSource) Results(_ context.Context, s domain.Settings, _ time.Time) ([]domain.Measurement, error) {
	f.token = s.Token
	return f.list, f.err
}

type rec struct{ got []kernel.Observation }

func (r *rec) Observe(_ context.Context, o kernel.Observation) { r.got = append(r.got, o) }

type clock struct{ t time.Time }

func (c clock) Now() time.Time { return c.t }

func TestSpeedtest(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	src := &fakeSource{list: []domain.Measurement{
		{At: now.Add(-48 * time.Hour), DownloadBps: 250e6, UploadBps: 40e6},
		{At: now.Add(-time.Hour), DownloadBps: 80e6, UploadBps: 40e6},
	}}
	repo, obs := &memRepo{}, &rec{}
	uc := NewSpeedtest(repo, src, nil, obs, clock{now})

	if _, err := uc.Results(ctx, "", 24); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unconfigured: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://nas:8765", Token: "secret", MinDownMbps: 100}); err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Settings(ctx)
	if !v.HasToken || v.MinDownMbps != 100 {
		t.Fatalf("%+v", v)
	}
	// An empty token keeps the stored one.
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://nas:8765", MinDownMbps: 100}); err != nil || src.token != "secret" {
		t.Fatalf("token kept: %v %q", err, src.token)
	}
	res, err := uc.Results(ctx, "", 24)
	if err != nil || len(res.History) != 1 || res.Latest.DownloadBps != 80e6 {
		t.Fatalf("%+v %v", res, err)
	}
	if week, _ := uc.Results(ctx, "", 0); len(week.History) != 2 {
		t.Fatalf("week: %+v", week)
	}
	if err := uc.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if len(obs.got) != 1 || obs.got[0].Healthy || obs.got[0].Kind != "speedtest" {
		t.Fatalf("observations: %+v", obs.got)
	}
	src.err = errors.New("down")
	_ = uc.Refresh(ctx)
	if res, _ := uc.Results(ctx, "", 24); res.Error == "" || res.Latest == nil {
		t.Fatalf("stale cache with error expected: %+v", res)
	}
}

func TestSaveRejectsUnreachable(t *testing.T) {
	uc := NewSpeedtest(&memRepo{}, &fakeSource{err: errors.New("401")}, nil, nil, clock{time.Now()})
	if _, err := uc.SaveSettings(context.Background(), SettingsInput{URL: "http://nas", Token: "x"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("%v", err)
	}
}

// conns resolves "office" to another Speedtest Tracker; everything else
// fails, like a missing copy or a connection of another service.
type conns struct{}

func (conns) Settings(_ context.Context, id string) (domain.Settings, error) {
	if id == "office" {
		return domain.Settings{URL: "http://office", Token: "office-token"}, nil
	}
	return domain.Settings{}, errors.New("no usable connection")
}

func TestResultsThroughConnection(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	src := &fakeSource{list: []domain.Measurement{
		{At: now.Add(-48 * time.Hour), DownloadBps: 250e6},
		{At: now.Add(-time.Hour), DownloadBps: 80e6},
	}}
	obs := &rec{}
	uc := NewSpeedtest(&memRepo{}, src, conns{}, obs, clock{now})
	res, err := uc.Results(ctx, "office", 24)
	if err != nil || src.token != "office-token" || len(res.History) != 1 || res.Latest == nil {
		t.Fatalf("office: %+v %v", res, err)
	}
	if _, err := uc.Results(ctx, "missing", 24); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("missing falls back to the (empty) module settings: %v", err)
	}
	src.token = ""
	if week, _ := uc.Results(ctx, "office", 0); len(week.History) != 2 || src.token != "" {
		t.Fatalf("a connection's results are cached and windowed: %+v", week)
	}
	if len(obs.got) != 0 {
		t.Fatal("a connection has no thresholds to observe")
	}
}
