package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/nextcloud/domain"
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
	uc := NewNextcloud(repo, src, nil, clk{}, nil)
	if _, err := uc.Status(ctx, ""); !errors.Is(err, kernel.ErrNotFound) {
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
	if v, _ := uc.Status(ctx, ""); v.Status.Version != "30.0.1" {
		t.Fatalf("status %+v", v)
	}
	src.err = errors.New("down")
	_ = uc.Refresh(ctx)
	if v, _ := uc.Status(ctx, ""); v.Error == "" || v.Status.Version != "30.0.1" {
		t.Fatalf("last known state must be kept with error: %+v", v)
	}
}

// conns resolves "office" to another Nextcloud; everything else fails,
// like a missing copy or a connection of another service.
type conns struct{}

func (conns) Settings(_ context.Context, id string) (domain.Settings, error) {
	if id == "office" {
		return domain.Settings{URL: "https://office", Username: "a", AppPassword: "p"}, nil
	}
	return domain.Settings{}, errors.New("no usable connection")
}

// countingSource counts the reads per address.
type countingSource struct {
	reads map[string]int
	err   error
}

func (c *countingSource) Status(_ context.Context, s domain.Settings) (domain.Status, error) {
	c.reads[s.URL]++
	return domain.Status{Version: s.URL}, c.err
}

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

func TestStatusThroughConnection(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{s: &domain.Settings{URL: "https://home", Username: "a", AppPassword: "p"}}
	src := &countingSource{reads: map[string]int{}}
	c := &movingClock{now: time.Unix(0, 0)}
	uc := NewNextcloud(repo, src, conns{}, c, nil)
	if v, err := uc.Status(ctx, "office"); err != nil || v.Status.Version != "https://office" {
		t.Fatalf("office: %+v %v", v, err)
	}
	if v, err := uc.Status(ctx, "missing"); err != nil || v.Status.Version != "https://home" {
		t.Fatalf("missing falls back to the module settings: %+v %v", v, err)
	}
	_, _ = uc.Status(ctx, "office")
	if src.reads["https://office"] != 1 {
		t.Fatal("a connection's status is cached")
	}
	c.now = c.now.Add(connectionTTL)
	src.err = errors.New("down")
	if v, _ := uc.Status(ctx, "office"); v.Error == "" || v.Status.Version != "https://office" || src.reads["https://office"] != 2 {
		t.Fatalf("a failed read keeps the last status: %+v", v)
	}
	repo.s = nil
	if _, err := uc.Status(ctx, "office"); err != nil {
		t.Fatalf("a connection works without module settings: %v", err)
	}
}
