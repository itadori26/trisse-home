package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeReleases struct {
	calls int
	err   error
}

func (f *fakeReleases) Latest(context.Context) (Release, bool, error) {
	f.calls++
	return Release{Version: "v1.3.0", URL: "https://example/releases/v1.3.0"}, true, f.err
}

type memUpdateSettings struct{ on bool }

func (m *memUpdateSettings) CheckEnabled(context.Context) (bool, error) { return m.on, nil }
func (m *memUpdateSettings) SetCheckEnabled(_ context.Context, on bool) error {
	m.on = on
	return nil
}

type stepClock struct{ t time.Time }

func (c *stepClock) Now() time.Time { return c.t }

func TestVersions(t *testing.T) {
	ctx := context.Background()
	src, set, clk := &fakeReleases{}, &memUpdateSettings{}, &stepClock{t: time.Unix(1_800_000_000, 0)}
	uc := NewVersions("1.2.0", src, set, clk)
	info, _ := uc.Info(ctx)
	if info.Current != "1.2.0" || info.Latest != nil || src.calls != 0 {
		t.Fatal("no outbound request while the check is disabled")
	}
	info, _ = uc.SetCheckEnabled(ctx, true)
	if !info.UpdateAvailable || info.Latest.Version != "v1.3.0" || src.calls != 1 {
		t.Fatalf("%+v", info)
	}
	_, _ = uc.Info(ctx)
	if src.calls != 1 {
		t.Fatal("cached result must be used within a day")
	}
	clk.t = clk.t.Add(25 * time.Hour)
	src.err = errors.New("rate limited")
	if info, err := uc.Info(ctx); err != nil || src.calls != 2 || !info.UpdateAvailable {
		t.Fatalf("stale lookup failure must keep the last result: %+v %v", info, err)
	}
	dev := NewVersions("dev", src, set, clk)
	if info, _ := dev.Info(ctx); info.UpdateAvailable {
		t.Fatal("development builds never show an update hint")
	}
}
