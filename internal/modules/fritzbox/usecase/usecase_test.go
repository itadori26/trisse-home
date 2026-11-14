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
	if f.callsErr != nil {
		return nil, f.callsErr
	}
	return []domain.Call{{Type: domain.CallMissed, Number: "0441 123"}, {Type: domain.CallOutgoing, Number: "0441 456"}}, nil
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
	all, _ := domain.NewCallFilter("", 0)
	if _, err := uc.Calls(ctx, all); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("calls without a box: %v", err)
	}
	// Without login: the status, and the calls widget explains why it is empty.
	if _, err := uc.SaveSettings(ctx, SettingsInput{}); err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Status(ctx)
	if !v.Status.WAN.Connected || v.Status.Model == "" {
		t.Fatalf("no-login status %+v", v)
	}
	if cv, err := uc.Calls(ctx, all); err != nil || len(cv.Calls) != 0 || cv.Hint == "" {
		t.Fatalf("no-login calls %+v %v", cv, err)
	}
	// With login: the background job never fetches calls (#168) …
	if _, err := uc.SaveSettings(ctx, SettingsInput{Username: "home", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	before := box.callRequests
	for range 3 {
		_ = uc.Refresh(ctx)
	}
	if box.callRequests != before {
		t.Fatalf("the background refresh fetched calls: %d→%d", before, box.callRequests)
	}
	// … a calls widget does, and the list is reused for two minutes.
	missed, _ := domain.NewCallFilter("missed", 5)
	cv, err := uc.Calls(ctx, all)
	if err != nil || len(cv.Calls) != 2 || box.callRequests != before+1 || cv.UpdatedAt.IsZero() {
		t.Fatalf("first fetch %+v %v (requests %d)", cv, err, box.callRequests)
	}
	if cv, _ := uc.Calls(ctx, missed); len(cv.Calls) != 1 || cv.Calls[0].Type != domain.CallMissed || box.callRequests != before+1 {
		t.Fatalf("filtered from the cache %+v (requests %d)", cv, box.callRequests)
	}
	c.t = c.t.Add(3 * time.Minute)
	_, _ = uc.Calls(ctx, all)
	if box.callRequests != before+2 {
		t.Fatal("calls not fetched again after two minutes")
	}
	// A failing box keeps the last list and says so.
	box.callsErr = errors.New("timeout")
	c.t = c.t.Add(3 * time.Minute)
	if cv, err := uc.Calls(ctx, all); err != nil || len(cv.Calls) != 2 || cv.Error == "" {
		t.Fatalf("stale calls on error %+v %v", cv, err)
	}
	box.callsErr = nil
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
