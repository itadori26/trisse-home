package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/media/domain"
	"trisse-home/internal/platform/database/dbtest"
)

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + strings.ToUpper(s), nil }
func (prefixCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return strings.ToLower(strings.TrimPrefix(s, "enc:")), nil
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "media", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s := domain.Settings{Provider: domain.ProviderPlex, URL: "http://plex:32400", Token: "plex-token"}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT token_encrypted FROM media_settings`).Scan(&stored)
	if strings.Contains(stored, "plex-token") {
		t.Fatalf("token stored in plain text: %q", stored)
	}
	if got, err := repo.Load(ctx); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

func TestJellyfin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != `MediaBrowser Token="key"` {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/System/Info":
			_, _ = w.Write([]byte(`{"ServerName":"Wohnzimmer","Version":"10.10.0"}`))
		case "/Items":
			if r.URL.Query().Get("SortBy") != "DateCreated" {
				t.Errorf("recently added must sort by date: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"Items":[{"Id":"abc","Name":"Pilot","SeriesName":"Show","Type":"Episode","ProductionYear":2026,"DateCreated":"2026-11-12T10:00:00Z"}]}`))
		case "/Sessions":
			_, _ = w.Write([]byte(`[{"UserName":"anna","DeviceName":"TV"},
				{"UserName":"ben","DeviceName":"Phone","NowPlayingItem":{"Name":"Film","Type":"Movie","RunTimeTicks":72000000000},
				 "PlayState":{"PositionTicks":36000000000,"IsPaused":true}}]`))
		case "/Items/abc/Images/Primary":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("jpeg"))
		case "/Items/svg/Images/Primary":
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = w.Write([]byte("<svg/>"))
		}
	}))
	defer srv.Close()
	jf := NewJellyfin(srv.Client())
	s := domain.Settings{Provider: domain.ProviderJellyfin, URL: srv.URL, Token: "key"}
	o, err := jf.Overview(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if o.ServerName != "Wohnzimmer" || len(o.Recent) != 1 || o.Recent[0].Series != "Show" || o.Recent[0].Kind != "episode" {
		t.Fatalf("recent: %+v", o)
	}
	if len(o.Sessions) != 1 || o.Sessions[0].User != "ben" || !o.Sessions[0].Paused || o.Sessions[0].Progress() != 50 {
		t.Fatalf("only sessions that play something: %+v", o.Sessions)
	}
	if p, err := jf.Poster(context.Background(), s, "abc"); err != nil || p.ContentType != "image/jpeg" {
		t.Fatalf("poster: %+v %v", p, err)
	}
	if _, err := jf.Poster(context.Background(), s, "svg"); err == nil {
		t.Fatal("SVG posters are rejected (scripts)")
	}
	s.Token = "bad"
	if _, err := jf.Overview(context.Background(), s); err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("expected token error, got %v", err)
	}
}

func TestPlex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Plex-Token") != "tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(`{"MediaContainer":{"friendlyName":"Plex"}}`))
		case "/library/recentlyAdded":
			_, _ = w.Write([]byte(`{"MediaContainer":{"Metadata":[{"ratingKey":"42","title":"Album","parentTitle":"Band","type":"album","addedAt":1794500000}]}}`))
		case "/status/sessions":
			_, _ = w.Write([]byte(`{"MediaContainer":{"Metadata":[{"title":"Ep","grandparentTitle":"Show","type":"episode","duration":1000,"viewOffset":250,
				"User":{"title":"anna"},"Player":{"title":"Shield","state":"playing"}}]}}`))
		case "/photo/:/transcode":
			if r.URL.Query().Get("url") != "/library/metadata/42/thumb" {
				t.Errorf("thumb url: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("jpeg"))
		}
	}))
	defer srv.Close()
	p := NewPlex(srv.Client())
	s := domain.Settings{Provider: domain.ProviderPlex, URL: srv.URL, Token: "tok"}
	o, err := p.Overview(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if o.ServerName != "Plex" || o.Recent[0].Series != "Band" || o.Recent[0].Kind != "album" || o.Recent[0].AddedAt != time.Unix(1794500000, 0).UTC() {
		t.Fatalf("recent: %+v", o.Recent)
	}
	if o.Sessions[0].Progress() != 25 || o.Sessions[0].Paused || o.Sessions[0].Device != "Shield" {
		t.Fatalf("sessions: %+v", o.Sessions)
	}
	if poster, err := p.Poster(context.Background(), s, "42"); err != nil || string(poster.Data) != "jpeg" {
		t.Fatalf("poster: %v", err)
	}
}
