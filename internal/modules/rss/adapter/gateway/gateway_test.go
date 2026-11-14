package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trisse-home/internal/modules/rss/domain"
	"trisse-home/internal/modules/rss/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

const atom = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Example Blog</title><link href="https://blog.example/"/>
  <entry><id>urn:1</id><title>First post</title><link href="https://blog.example/1"/><updated>2026-11-01T10:00:00Z</updated></entry>
  <entry><id>urn:2</id><title>Second &amp; last</title><link href="https://blog.example/2"/><updated>2026-11-02T10:00:00Z</updated></entry>
</feed>`

func TestHTTPFetcher(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/feed":
			_, _ = w.Write([]byte(atom))
		case "/html":
			_, _ = w.Write([]byte("<html><body>nope</body></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	f := NewHTTPFetcher(srv.Client())
	got, err := f.Fetch(context.Background(), srv.URL+"/feed")
	if err != nil || got.Title != "Example Blog" || len(got.Items) != 2 || got.Items[1].Title != "Second & last" {
		t.Fatalf("unexpected %+v %v", got, err)
	}
	if got.Items[0].Published.IsZero() {
		t.Fatal("published date not parsed")
	}
	for _, p := range []string{"/html", "/missing"} {
		if _, err := f.Fetch(context.Background(), srv.URL+p); err == nil {
			t.Errorf("%s: expected error", p)
		}
	}
}

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "rss", Migrations()))
	now := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	feed, _ := domain.NewFeed("f1", "https://blog.example/feed", now)
	if err := repo.CreateFeed(ctx, feed); err != nil {
		t.Fatal(err)
	}
	dup, _ := domain.NewFeed("f2", "https://blog.example/feed", now)
	if err := repo.CreateFeed(ctx, dup); !errors.Is(err, domain.ErrAlreadySubscribed) {
		t.Fatalf("expected duplicate, got %v", err)
	}
	items := []domain.Item{
		domain.NewItem("i1", "f1", "g1", "old", "", now.Add(-2*time.Hour)),
		domain.NewItem("i2", "f1", "g2", "new", "", now.Add(-time.Hour)),
		domain.NewItem("i3", "f1", "g3", "newest", "", now),
	}
	if n, err := repo.AddItems(ctx, items); err != nil || n != 3 {
		t.Fatalf("add: %d %v", n, err)
	}
	if err := repo.SetRead(ctx, "i3", true); err != nil {
		t.Fatal(err)
	}
	// Re-adding a known GUID must keep the read flag.
	_, _ = repo.AddItems(ctx, []domain.Item{domain.NewItem("x", "f1", "g3", "newest (edited)", "", now)})
	list, _ := repo.ListItems(ctx, usecase.ItemFilter{Limit: 10})
	if len(list) != 3 || list[0].ID != "i3" || !list[0].Read || list[0].Title != "newest (edited)" {
		t.Fatalf("unexpected list %+v", list)
	}
	if err := repo.PruneItems(ctx, "f1", 2); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.ListItems(ctx, usecase.ItemFilter{Limit: 10}); len(list) != 2 || list[1].ID != "i2" {
		t.Fatalf("prune kept wrong items: %+v", list)
	}
	counts, _ := repo.UnreadCounts(ctx)
	if counts["f1"] != 1 {
		t.Fatalf("unread count %v", counts)
	}
	if err := repo.DeleteFeed(ctx, "f1"); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.ListItems(ctx, usecase.ItemFilter{Limit: 10}); len(list) != 0 {
		t.Fatal("items not cascaded")
	}
}

func TestRepositoryKeepsCategory(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "rss", Migrations()))
	f, _ := domain.NewFeed("c", "https://a.b/feed", time.Now())
	f.Category = "Tech / Go"
	if err := repo.CreateFeed(ctx, f); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.GetFeed(ctx, "c"); got.Category != "Tech / Go" {
		t.Fatalf("%+v", got)
	}
}
