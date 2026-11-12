package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/speedtest/domain"
	"trisse-home/internal/platform/database/dbtest"
)

type rot struct{}

func (rot) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (rot) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "speedtest", Migrations())
	r := NewSettingsRepository(db, rot{})
	if _, err := r.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatal(err)
	}
	_ = r.Save(ctx, domain.Settings{URL: "http://nas", Token: "tok", MinDownMbps: 100})
	var stored string
	_ = db.QueryRow(`SELECT token_encrypted FROM speedtest_settings`).Scan(&stored)
	s, err := r.Load(ctx)
	if err != nil || s.Token != "tok" || s.MinDownMbps != 100 || stored != "enc:tok" {
		t.Fatalf("%+v %v %q", s, err, stored)
	}
}

const body = `{"data":[
 {"id":3,"ping":12.5,"download":31250000,"upload":5000000,"download_bits":250000000,"upload_bits":40000000,"status":"completed","created_at":"2026-11-12T10:00:00.000000Z"},
 {"id":2,"ping":0,"download":0,"upload":0,"status":"failed","created_at":"2026-11-12T09:00:00Z"},
 {"id":1,"ping":14,"download":12500000,"upload":2500000,"status":"completed","created_at":"2026-11-12 08:00:00"}
],"links":{},"meta":{}}`

func TestAPIResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" || r.URL.Path != "/api/v1/results" ||
			r.URL.Query().Get("filter[start_at]") != "2026-11-11" || r.Header.Get("Accept") != "application/json" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	since := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	list, err := NewAPI(srv.Client()).Results(context.Background(), domain.Settings{URL: srv.URL, Token: "tok"}, since)
	if err != nil || len(list) != 2 {
		t.Fatalf("%+v %v", list, err)
	}
	if list[0].DownloadBps != 100e6 || list[1].DownloadBps != 250e6 || list[1].PingMs != 12.5 || !list[0].At.Before(list[1].At) {
		t.Fatalf("%+v", list)
	}
	if _, err := NewAPI(srv.Client()).Results(context.Background(), domain.Settings{URL: srv.URL, Token: "bad"}, since); err == nil {
		t.Fatal("401 must fail")
	}
}
