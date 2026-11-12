package gateway

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/platform/database/dbtest"
)

// pngHeader is enough for content sniffing to report image/png.
var pngHeader = []byte("\x89PNG\r\n\x1a\n0000IHDR")

func TestParseHead(t *testing.T) {
	doc := `<!doctype html><html><head><meta charset="utf-8">
		<TITLE>Go &amp; Friends
		</TITLE>
		<link rel="icon" type="image/svg+xml" href="/icon.svg">
		<link rel="apple-touch-icon" href="/touch.png">
		<link href='/fav.png' rel='icon' type='image/png'>
		</head><body><link rel="icon" href="/body.png"><title>ignored</title></body></html>`
	meta := ParseHead(doc)
	if meta.Title != "Go & Friends\n\t\t" || meta.IconURL != "/fav.png" {
		t.Fatalf("got %+v", meta)
	}
	if got := ParseHead(`<head><link rel="apple-touch-icon" href="t.png"></head>`); got.IconURL != "t.png" || got.Title != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestWebFetchesMetadataIconsAndChecksLinks(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/page", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<head><title>Docs</title><link rel="icon" href="img/i.png"></head>`))
	})
	mux.HandleFunc("/img/i.png", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml") // lies: content wins
		_, _ = w.Write(pngHeader)
	})
	mux.HandleFunc("/evil.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	})
	mux.HandleFunc("/big.png", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(append(pngHeader, bytes.Repeat([]byte{0}, domain.MaxIconBytes)...))
	})
	mux.HandleFunc("/json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/page", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/no-head", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx := context.Background()
	// Tests run against loopback, so the guard of the production client is not used here.
	web := NewWeb(&http.Client{Timeout: 5 * time.Second})

	meta, err := web.Fetch(ctx, srv.URL+"/page")
	if err != nil || meta.Title != "Docs" || meta.IconURL != srv.URL+"/img/i.png" {
		t.Fatalf("fetch %+v %v", meta, err)
	}
	if _, err := web.Fetch(ctx, srv.URL+"/json"); err == nil {
		t.Fatal("non-HTML pages must be refused")
	}
	icon, err := web.FetchIcon(ctx, srv.URL+"/img/i.png")
	if err != nil || icon.ContentType != "image/png" {
		t.Fatalf("icon %+v %v", icon.ContentType, err)
	}
	if _, err := web.FetchIcon(ctx, srv.URL+"/evil.svg"); err == nil {
		t.Fatal("SVG icons must be refused")
	}
	if _, err := web.FetchIcon(ctx, srv.URL+"/big.png"); err == nil {
		t.Fatal("oversized icons must be refused")
	}

	if code, final, failed := web.Check(ctx, srv.URL+"/page"); failed || code != 200 || final != srv.URL+"/page" {
		t.Fatalf("check %d %s %v", code, final, failed)
	}
	if code, final, _ := web.Check(ctx, srv.URL+"/old"); code != 200 || !strings.HasSuffix(final, "/page") {
		t.Fatalf("redirect %d %s", code, final)
	}
	if code, _, _ := web.Check(ctx, srv.URL+"/no-head"); code != 200 {
		t.Fatalf("GET fallback %d", code)
	}
	if code, _, _ := web.Check(ctx, srv.URL+"/missing"); code != 404 {
		t.Fatalf("missing %d", code)
	}
	if _, _, failed := web.Check(ctx, "http://127.0.0.1:1/"); !failed {
		t.Fatal("unreachable host must fail")
	}
}

func TestLinkStatusIconAndSettingsStorage(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "bookmarks", Migrations()))
	if s, err := repo.LoadSettings(ctx); err != nil || s.Enrich || s.CheckLinks {
		t.Fatalf("defaults must be off: %+v %v", s, err)
	}
	if err := repo.SaveSettings(ctx, domain.Settings{Enrich: true, CheckLinks: true}); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.LoadSettings(ctx); !s.Enrich || !s.CheckLinks {
		t.Fatalf("settings %+v", s)
	}

	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	b, _ := domain.NewBookmark("b1", domain.BookmarkInput{URL: "https://go.dev"}, now)
	if err := repo.SaveBookmark(ctx, b); err != nil {
		t.Fatal(err)
	}
	status := domain.LinkStatus{State: domain.LinkRedirect, Code: 200, Target: "https://go.dev/doc/", CheckedAt: now}
	if err := repo.SaveLinkStatus(ctx, "b1", status); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.ListLinkStatuses(ctx); got["b1"] != status {
		t.Fatalf("status %+v", got["b1"])
	}
	if _, err := repo.Icon(ctx, "b1"); !errors.Is(err, domain.ErrIconNotFound) {
		t.Fatalf("missing icon: %v", err)
	}
	icon, _ := domain.NewIcon("image/png", pngHeader)
	if err := repo.SaveIcon(ctx, "b1", icon); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.Icon(ctx, "b1"); err != nil || !bytes.Equal(got.Data, pngHeader) || got.ContentType != "image/png" {
		t.Fatalf("icon %+v %v", got, err)
	}
	if ids, _ := repo.IconIDs(ctx); !ids["b1"] {
		t.Fatal("icon id missing")
	}
	// Deleting the bookmark removes its status and icon.
	if err := repo.DeleteBookmark(ctx, "b1"); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.ListLinkStatuses(ctx); len(got) != 0 {
		t.Fatalf("status not removed: %+v", got)
	}
	if ids, _ := repo.IconIDs(ctx); len(ids) != 0 {
		t.Fatal("icon not removed")
	}
}
