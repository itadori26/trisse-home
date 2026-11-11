package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/platform/database/dbtest"
)

func TestGitHubReleases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/none" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"tag_name":"v1.4.0","html_url":"https://github.com/x/releases/tag/v1.4.0","published_at":"2026-11-11T10:00:00Z"}`)
	}))
	defer srv.Close()
	rel, ok, err := NewGitHubReleases(srv.Client(), srv.URL+"/latest").Latest(context.Background())
	if err != nil || !ok || rel.Version != "v1.4.0" || rel.PublishedAt.IsZero() {
		t.Fatalf("%+v %v %v", rel, ok, err)
	}
	if _, ok, err := NewGitHubReleases(srv.Client(), srv.URL+"/none").Latest(context.Background()); ok || err != nil {
		t.Fatal("no release yet must not be an error")
	}
}

func TestUpdateSettings(t *testing.T) {
	ctx := context.Background()
	s := NewUpdateSettings(dbtest.Open(t, "system", Migrations()))
	if on, err := s.CheckEnabled(ctx); on || err != nil {
		t.Fatal("check must be off by default")
	}
	_ = s.SetCheckEnabled(ctx, true)
	if on, _ := s.CheckEnabled(ctx); !on {
		t.Fatal("not stored")
	}
}
