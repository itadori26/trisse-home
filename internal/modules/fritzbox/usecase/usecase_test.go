package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/fritzbox/domain"
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

type fakeBox struct {
	wanErr, callsErr error
	callRequests     int
}

func (f *fakeBox) WAN(context.Context, domain.Settings) (domain.WAN, string, error) {
	return domain.WAN{Connected: true, Known: true, ExternalIP: "203.0.113.7"}, "FRITZ!Box 6670 Cable", f.wanErr
}

func (f *fakeBox) Calls(_ context.Context, s domain.Settings, _ int) ([]domain.Call, error) {
	f.callRequests++
	if s.Password != "pw" {
		return nil, errors.New("401")
	}
	return []domain.Call{{Type: domain.CallMissed, Number: "0441 123"}}, f.callsErr
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

func TestFritzBox(t *testing.T) {
	ctx := context.Background()
	repo, box, c := &memRepo{}, &fakeBox{}, &clk{t: time.Unix(1_800_000_000, 0)}
	uc := NewFritzBox(repo, box, c)
	if v, _ := uc.Settings(ctx); v.URL != domain.DefaultURL || v.Configured {
		t.Fatalf("defaults %+v", v)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{Username: "home", Password: "wrong"}); !errors.Is(err, kernel.ErrInvalid) || repo.s != nil {
		t.Fatalf("wrong login must not be saved: %v", err)
	}
	// Without login: status but no calls.
	if _, err := uc.SaveSettings(ctx, SettingsInput{}); err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Status(ctx)
	if !v.Status.WAN.Connected || len(v.Status.Calls) != 0 || v.Status.Hint == "" || v.Status.Model == "" {
		t.Fatalf("no-login status %+v", v)
	}
	// With login: calls, fetched at most every two minutes.
	if _, err := uc.SaveSettings(ctx, SettingsInput{Username: "home", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	before := box.callRequests
	_ = uc.Refresh(ctx)
	if v, _ := uc.Status(ctx); len(v.Status.Calls) != 1 || box.callRequests != before {
		t.Fatalf("calls %+v, requests %d→%d", v.Status.Calls, before, box.callRequests)
	}
	c.t = c.t.Add(3 * time.Minute)
	_ = uc.Refresh(ctx)
	if box.callRequests != before+1 {
		t.Fatal("calls not refreshed after two minutes")
	}
	// The stored password is kept for the same box (default port added) …
	stored := repo.s.URL
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: strings.TrimSuffix(stored, ":49000"), Username: "home"}); err != nil || repo.s.Password != "pw" {
		t.Fatalf("password not kept for %s: %v", stored, err)
	}
	// … but never sent to another address (#191).
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://192.168.1.1", Username: "home"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("password reused for a new address: %v", err)
	}
	box.wanErr = errors.New("down")
	_ = uc.Refresh(ctx)
	if v, _ := uc.Status(ctx); v.Error == "" {
		t.Fatal("error not reported")
	}
}
