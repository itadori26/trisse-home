package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/nextcloud/domain"
	"trisse-home/internal/shared/kernel"
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
	err  error
	seen domain.Settings
}

func (f *fakeSource) Status(_ context.Context, s domain.Settings) (domain.Status, error) {
	f.seen = s
	return domain.Status{Version: "30.0.1"}, f.err
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestNextcloud(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{}
	uc := NewNextcloud(repo, src, clk{})
	if _, err := uc.Status(ctx); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unconfigured: %v", err)
	}
	src.err = errors.New("401")
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "https://c", Username: "a", AppPassword: "p"}); err == nil || repo.s != nil {
		t.Fatal("failed connection must not be saved")
	}
	src.err = nil
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "https://c", Username: "a", AppPassword: "p"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "https://c", Username: "a"}); err != nil || src.seen.AppPassword != "p" {
		t.Fatalf("empty password must keep the stored one: %v", err)
	}
	if v, _ := uc.Status(ctx); v.Status.Version != "30.0.1" {
		t.Fatalf("status %+v", v)
	}
	src.err = errors.New("down")
	_ = uc.Refresh(ctx)
	if v, _ := uc.Status(ctx); v.Error == "" || v.Status.Version != "30.0.1" {
		t.Fatalf("last known state must be kept with error: %+v", v)
	}
}
