package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "services", Migrations()))
	s, _ := domain.NewService("1", "NAS", "https://nas.lan", "", "", "", true, 0)
	if err := repo.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, "1")
	if err != nil || !got.SkipTLSVerify || got.CheckURL != "https://nas.lan" {
		t.Fatalf("round trip: %+v %v", got, err)
	}
	if err := repo.Delete(ctx, "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, "1"); !errors.Is(err, domain.ErrServiceNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestHTTPChecker(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	c := NewHTTPChecker(http.DefaultClient, srv.Client())
	if r := c.Check(context.Background(), srv.URL, false); r.Err == nil {
		t.Fatal("self-signed certificate accepted without opt-in")
	}
	r := c.Check(context.Background(), srv.URL, true)
	if r.Err != nil || r.HTTPStatus != http.StatusUnauthorized || r.CertExpiresAt.IsZero() {
		t.Fatalf("unexpected %+v", r)
	}
}

func TestHistoryStore(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "services", Migrations())
	repo, hist := NewRepository(db), NewHistoryStore(db)
	s, _ := domain.NewService("s1", "NAS", "https://nas.lan", "", "", "", false, 0)
	_ = repo.Create(ctx, s)
	t0 := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	for i := range 3 {
		if err := hist.Add(ctx, domain.LatencyPoint{ServiceID: "s1", At: t0.Add(time.Duration(i) * time.Minute), LatencyMs: 10, Online: i != 1}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := hist.Since(ctx, t0)
	if err != nil || len(got) != 3 || got[1].Online {
		t.Fatalf("since %+v %v", got, err)
	}
	_ = hist.PruneBefore(ctx, t0.Add(time.Minute))
	if got, _ := hist.Since(ctx, t0); len(got) != 2 {
		t.Fatalf("prune: %d", len(got))
	}
	_ = repo.Delete(ctx, "s1")
	if got, _ := hist.Since(ctx, t0); len(got) != 0 {
		t.Fatal("history not removed with the service")
	}
}
