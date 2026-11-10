package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/adguard/domain"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeSource struct{ err error }

func (f *fakeSource) Stats(_ context.Context, s domain.Settings) (domain.Stats, error) {
	if f.err != nil {
		return domain.Stats{}, f.err
	}
	if s.Password != "secret" {
		return domain.Stats{}, errors.New("401 Unauthorized")
	}
	return domain.Stats{Queries: 100, Blocked: 20, AvgProcessingTime: 1500 * time.Microsecond}, nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestAdGuard(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{}
	uc := NewAdGuard(repo, src, clk{})

	if s, _ := uc.Settings(ctx); s.Configured {
		t.Fatal("expected unconfigured")
	}
	if _, err := uc.Stats(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ag:3000", Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("expected connection test to fail")
	}
	if repo.s != nil {
		t.Fatal("invalid settings must not be stored")
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ag:3000", Username: "admin", Password: "secret"}); err != nil {
		t.Fatal(err)
	}
	// An empty password keeps the stored one.
	view, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ag:3000", Username: "admin2"})
	if err != nil || !view.HasPassword || repo.s.Password != "secret" {
		t.Fatalf("password not kept: %+v %v", view, err)
	}
	stats, _ := uc.Stats(ctx)
	if stats.BlockedPercent != 20 || stats.AvgProcessingMs != 1.5 {
		t.Fatalf("stats %+v", stats)
	}
	src.err = errors.New("timeout")
	_ = uc.Refresh(ctx)
	if stats, _ := uc.Stats(ctx); stats.Error == "" || stats.Queries != 100 {
		t.Fatalf("last known stats not kept: %+v", stats)
	}
}
