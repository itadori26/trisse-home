package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

func TestDHL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("DHL-API-Key") != "key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Query().Get("trackingNumber") == "UNKNOWN123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, `{"shipments":[{"id":"00340434","status":{"timestamp":"2026-11-11T14:02:00","statusCode":"transit",
			"status":"Die Sendung wurde im Zustellfahrzeug geladen","location":{"address":{"addressLocality":"Oldenburg"}}},
			"estimatedTimeOfDeliveryTimeFrame":{"from":"2026-11-12T09:00:00+01:00","to":"2026-11-12T13:00:00+01:00"}}]}`)
	}))
	defer srv.Close()
	d := NewDHL(srv.Client(), srv.URL)
	d.interval = time.Millisecond
	ctx := context.Background()
	st, est, err := d.Track(ctx, domain.Settings{DHLAPIKey: "key"}, "00340434")
	if err != nil || st.Code != domain.StatusTransit || st.Location != "Oldenburg" || !strings.Contains(st.Description, "Zustellfahrzeug") || est.To == 0 {
		t.Fatalf("%+v %+v %v", st, est, err)
	}
	if _, _, err := d.Track(ctx, domain.Settings{DHLAPIKey: "bad"}, "00340434"); !errors.Is(err, kernel.ErrUnauthorized) {
		t.Fatalf("bad key: %v", err)
	}
	if _, _, err := d.Track(ctx, domain.Settings{DHLAPIKey: "key"}, "UNKNOWN123"); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unknown: %v", err)
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "parcels", Migrations()), prefixCipher{})
	if s, err := repo.LoadSettings(ctx); err != nil || s.HideDeliveredDays != domain.DefaultHideDeliveredDays {
		t.Fatalf("defaults %+v %v", s, err)
	}
	_ = repo.SaveSettings(ctx, domain.Settings{DHLAPIKey: "k", HideDeliveredDays: 5})
	if s, _ := repo.LoadSettings(ctx); s.DHLAPIKey != "k" || s.HideDeliveredDays != 5 {
		t.Fatalf("%+v", s)
	}
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	p, _ := domain.NewParcel("p1", "00340434123456", domain.CarrierDHL, "Shoes", now)
	p.Status = domain.Status{Code: domain.StatusTransit, Description: "x", At: now}
	if err := repo.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, "p1")
	if err != nil || got.Status.Code != domain.StatusTransit || !got.Status.At.Equal(now) || got.Label != "Shoes" {
		t.Fatalf("%+v %v", got, err)
	}
	_ = repo.Delete(ctx, "p1")
	if _, err := repo.Get(ctx, "p1"); !errors.Is(err, domain.ErrParcelNotFound) {
		t.Fatal("not deleted")
	}
}
