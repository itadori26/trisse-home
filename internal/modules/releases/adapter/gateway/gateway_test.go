package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/platform/database/dbtest"
)

const radiohead = "a74b1b7f-71a5-4011-9441-d0b5e4122711"

func TestArtistRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewArtistRepository(dbtest.Open(t, "releases", Migrations()))
	for _, a := range []domain.Artist{{ID: "b", Name: "B"}, {ID: "a", Name: "A"}, {ID: "b", Name: "B again"}} {
		if err := repo.Add(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0].ID != "b" || list[0].Name != "B" {
		t.Fatalf("insertion order, idempotent: %+v %v", list, err)
	}
	_ = repo.Remove(ctx, "b")
	if list, _ := repo.List(ctx); len(list) != 1 {
		t.Fatalf("remove: %+v", list)
	}
}

func TestMusicBrainz(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ws/2/artist":
			_, _ = w.Write([]byte(`{"artists":[{"id":"` + radiohead + `","name":"Radiohead","disambiguation":"UK band"}]}`))
		case "/ws/2/release-group":
			if q := r.URL.Query().Get("query"); q != "arid:"+radiohead+" AND firstreleasedate:[2026-08-01 TO *]" {
				t.Errorf("query: %q", q)
			}
			_, _ = w.Write([]byte(`{"release-groups":[{"id":"r1","title":"New LP","primary-type":"Album","first-release-date":"2026-11-07"},
				{"id":"r2","title":"Old","primary-type":"Single","first-release-date":"2020"},
				{"id":"r3","title":"Undated","first-release-date":""},
				{"id":"r4","title":"Month only","first-release-date":"2026-09"}]}`))
		case "/cover/release-group/r1/front-250":
			http.Redirect(w, r, "/img/r1.jpg", http.StatusTemporaryRedirect)
		case "/img/r1.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("jpeg"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	mb := NewMusicBrainz(srv.Client(), srv.URL+"/ws/2", srv.URL+"/cover")
	var waited []time.Duration
	clock := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	mb.now = func() time.Time { return clock }
	mb.wait = func(_ context.Context, d time.Duration) error {
		waited = append(waited, d)
		clock = clock.Add(d)
		return nil
	}
	ctx := context.Background()

	artists, err := mb.SearchArtists(ctx, "radiohead")
	if err != nil || len(artists) != 1 || artists[0].Disambiguation != "UK band" {
		t.Fatalf("search: %+v %v", artists, err)
	}
	list, err := mb.ReleasesSince(ctx, domain.Artist{ID: radiohead, Name: "Radiohead"}, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || len(list) != 2 || list[0].Title != "New LP" || list[1].Type != "Other" || list[0].Artist != "Radiohead" {
		t.Fatalf("releases within the period, undated ones skipped: %+v %v", list, err)
	}
	if len(waited) != 1 || waited[0] != minInterval {
		t.Fatalf("the second request waits for the rate limit: %v", waited)
	}
	if c, err := mb.Cover(ctx, "r1"); err != nil || string(c.Data) != "jpeg" {
		t.Fatalf("cover follows the redirect: %+v %v", c, err)
	}
	if _, err := mb.Cover(ctx, "missing"); err == nil || !strings.Contains(err.Error(), "no cover") {
		t.Fatalf("missing cover: %v", err)
	}
}
