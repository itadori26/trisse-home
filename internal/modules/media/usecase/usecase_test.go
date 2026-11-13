package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/media/domain"
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

type fakeServer struct {
	err     error
	calls   int
	posters []string
}

func (f *fakeServer) Overview(_ context.Context, s domain.Settings) (domain.Overview, error) {
	if f.err != nil {
		return domain.Overview{}, f.err
	}
	if s.Token != "key" {
		return domain.Overview{}, errors.New("the API key or token was rejected")
	}
	f.calls++
	o := domain.Overview{ServerName: "Home"}
	for i := range domain.MaxRecent + 3 {
		o.Recent = append(o.Recent, domain.Item{ID: string(rune('a' + i)), Title: "x"})
	}
	return o, nil
}

func (f *fakeServer) Poster(_ context.Context, _ domain.Settings, id string) (domain.Poster, error) {
	f.posters = append(f.posters, id)
	return domain.Poster{ContentType: "image/jpeg", Data: []byte{1}}, nil
}

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

func setup() (Media, *memRepo, *fakeServer, *fakeServer, *movingClock) {
	repo, jf, plex := &memRepo{}, &fakeServer{}, &fakeServer{}
	c := &movingClock{now: time.Unix(0, 0)}
	uc := NewMedia(repo, map[domain.Provider]MediaServerProvider{domain.ProviderJellyfin: jf, domain.ProviderPlex: plex}, c)
	return uc, repo, jf, plex, c
}

func TestSettings(t *testing.T) {
	ctx := context.Background()
	uc, repo, _, plex, _ := setup()
	if s, _ := uc.Settings(ctx); s.Configured || s.Provider != "jellyfin" {
		t.Fatalf("defaults: %+v", s)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{Provider: "jellyfin", URL: "http://jf", Token: "bad"}); !errors.Is(err, kernel.ErrInvalid) || repo.s != nil {
		t.Fatalf("a failing connection test must not save: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{Provider: "jellyfin", URL: "http://jf", Token: "key"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{Provider: "jellyfin", URL: "http://jf2"}); err != nil || repo.s.Token != "key" {
		t.Fatalf("an empty token keeps the stored one: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{Provider: "plex", URL: "http://plex"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("switching the provider needs a new token: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{Provider: "plex", URL: "http://plex", Token: "key"}); err != nil || plex.calls != 1 {
		t.Fatalf("plex: %v %d", err, plex.calls)
	}
}

func TestOverviewAndPoster(t *testing.T) {
	ctx := context.Background()
	uc, repo, jf, _, c := setup()
	if _, err := uc.Overview(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
	repo.s = &domain.Settings{Provider: domain.ProviderJellyfin, URL: "http://jf", Token: "key"}
	o, err := uc.Overview(ctx)
	if err != nil || len(o.Recent) != domain.MaxRecent || o.ServerName != "Home" {
		t.Fatalf("overview is capped: %d %v", len(o.Recent), err)
	}
	_, _ = uc.Overview(ctx)
	if jf.calls != 1 {
		t.Fatalf("cached within the TTL, got %d calls", jf.calls)
	}
	c.now = c.now.Add(overviewTTL)
	_, _ = uc.Overview(ctx)
	if jf.calls != 2 {
		t.Fatalf("fetched again after the TTL, got %d calls", jf.calls)
	}

	if _, err := uc.Poster(ctx, "a"); err != nil || len(jf.posters) != 1 {
		t.Fatalf("poster of a recent item: %v", err)
	}
	for _, id := range []string{"zz", "../System/Info"} {
		if _, err := uc.Poster(ctx, id); !errors.Is(err, domain.ErrPosterNotFound) {
			t.Errorf("%q: only IDs of the overview are proxied, got %v", id, err)
		}
	}
	if len(jf.posters) != 1 {
		t.Fatalf("unknown IDs must not reach the server: %v", jf.posters)
	}
	jf.err = errors.New("down")
	c.now = c.now.Add(overviewTTL)
	if _, err := uc.Overview(ctx); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("unreachable: %v", err)
	}
}
