package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	if _, _, err := c.Check(context.Background(), srv.URL, false); err == nil {
		t.Fatal("self-signed certificate accepted without opt-in")
	}
	code, _, err := c.Check(context.Background(), srv.URL, true)
	if err != nil || code != http.StatusUnauthorized {
		t.Fatalf("code %d err %v", code, err)
	}
}
