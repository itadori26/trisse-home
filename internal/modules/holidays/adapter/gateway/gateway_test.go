package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestICSSource(t *testing.T) {
	data, err := os.ReadFile("testdata/ferien-ni.ics")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(data) }))
	defer srv.Close()
	loc, _ := time.LoadLocation("Europe/Berlin")
	from, to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := NewICSSource(srv.Client(), loc).Fetch(context.Background(), srv.URL, from, to)
	if err != nil || len(got) != 4 {
		t.Fatalf("%+v %v", got, err)
	}
	want := []string{"2026-10-12..2026-10-23", "2026-12-23..2027-01-05", "2027-02-01..2027-02-02", "2027-05-07..2027-05-07"}
	for i, p := range got {
		if s := p.Start.Format(time.DateOnly) + ".." + p.End.Format(time.DateOnly); s != want[i] {
			t.Errorf("%d: %s (%s), want %s", i, s, p.Name, want[i])
		}
	}
	if got[1].Name != "Weihnachtsferien 2026/2027 Niedersachsen" {
		t.Fatalf("name: %q", got[1].Name)
	}
	bad := httptest.NewServer(http.NotFoundHandler())
	defer bad.Close()
	if _, err := NewICSSource(bad.Client(), loc).Fetch(context.Background(), bad.URL, from, to); err == nil {
		t.Fatal("404 must be an error")
	}
}

func TestStore(t *testing.T) {
	ctx := context.Background()
	store := NewStore(dbtest.Open(t, "holidays", Migrations()))
	if s, err := store.Load(ctx); err != nil || s.SchoolURL != "" {
		t.Fatalf("%+v %v", s, err)
	}
	_ = store.Save(ctx, domain.Settings{SchoolURL: "https://x/ferien.ics"})
	if s, _ := store.Load(ctx); s.SchoolURL != "https://x/ferien.ics" {
		t.Fatalf("%+v", s)
	}
	cache := NewCache(store)
	if p, url, at, err := cache.Load(ctx); err != nil || p != nil || url != "" || !at.IsZero() {
		t.Fatalf("empty cache: %v %q %v %v", p, url, at, err)
	}
	d := func(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }
	at := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	periods := []domain.SchoolHoliday{{Name: "Weihnachtsferien", Start: d(2026, 12, 23), End: d(2027, 1, 5)}, {Name: "Herbstferien", Start: d(2026, 10, 12), End: d(2026, 10, 23)}}
	if err := cache.Store(ctx, "https://x/ferien.ics", periods, at); err != nil {
		t.Fatal(err)
	}
	p, url, gotAt, err := cache.Load(ctx)
	if err != nil || len(p) != 2 || p[0].Name != "Herbstferien" || !p[1].End.Equal(d(2027, 1, 5)) || url != "https://x/ferien.ics" || !gotAt.Equal(at) {
		t.Fatalf("%+v %q %v %v", p, url, gotAt, err)
	}
	_ = cache.Store(ctx, "https://x/ferien.ics", periods[:1], at)
	if p, _, _, _ := cache.Load(ctx); len(p) != 1 {
		t.Fatalf("store replaces: %+v", p)
	}
	_ = cache.Clear(ctx)
	if p, _, _, _ := cache.Load(ctx); len(p) != 0 {
		t.Fatal("clear")
	}
}
