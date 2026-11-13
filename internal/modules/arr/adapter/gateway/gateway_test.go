package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/arr/domain"
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

func TestInstanceRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "arr", Migrations())
	repo := NewInstanceRepository(db, prefixCipher{})
	a := domain.Instance{ID: "a", App: domain.AppSonarr, Name: "TV", URL: "http://s", APIKey: "secretkey"}
	b := domain.Instance{ID: "b", App: domain.AppProwlarr, Name: "Indexers", URL: "http://p", APIKey: "other"}
	for _, in := range []domain.Instance{a, b} {
		if err := repo.Save(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT api_key_encrypted FROM arr_instances WHERE id = 'a'`).Scan(&stored)
	if strings.Contains(stored, "secretkey") {
		t.Fatalf("API key stored in plain text: %q", stored)
	}
	a.Name = "Series"
	_ = repo.Save(ctx, a)
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0] != a || list[1] != b {
		t.Fatalf("list keeps creation order: %+v %v", list, err)
	}
	_ = repo.Delete(ctx, "a")
	if _, err := repo.Get(ctx, "a"); !errors.Is(err, domain.ErrInstanceNotFound) {
		t.Fatalf("deleted: %v", err)
	}
}

func arrServer(t *testing.T, app string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/v3/system/status", "/api/v1/system/status":
			_, _ = w.Write([]byte(`{"appName":"` + app + `","version":"4.0"}`))
		case "/api/v3/calendar":
			if app == "Sonarr" {
				_, _ = w.Write([]byte(`[{"title":"Pilot","seasonNumber":1,"episodeNumber":2,"airDateUtc":"2026-11-13T02:00:00Z","series":{"title":"Show"}}]`))
			} else {
				_, _ = w.Write([]byte(`[{"title":"Film","year":2026,"inCinemas":"2025-01-01T00:00:00Z","digitalRelease":"2026-11-20T00:00:00Z","hasFile":false},
					{"title":"Old","year":2020,"inCinemas":"2020-01-01T00:00:00Z"}]`))
			}
		case "/api/v3/queue":
			_, _ = w.Write([]byte(`{"records":[{"title":"Show.S01E01","status":"downloading","size":1000,"sizeleft":250,"timeleft":"1.02:03:04"},
				{"title":"Broken","status":"completed","trackedDownloadStatus":"warning","size":10,"sizeleft":0,"statusMessages":[{"messages":["No files found"]}]}]}`))
		case "/api/v3/health", "/api/v1/health":
			_, _ = w.Write([]byte(`[{"type":"warning","message":"Indexer unavailable"}]`))
		case "/api/v3/wanted/missing":
			_, _ = w.Write([]byte(`{"totalRecords":7}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestSonarr(t *testing.T) {
	srv := arrServer(t, "Sonarr")
	defer srv.Close()
	s := NewSonarr(srv.Client())
	in := domain.Instance{App: domain.AppSonarr, Name: "TV", URL: srv.URL, APIKey: "key"}
	if err := s.Check(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	snap, err := s.Snapshot(context.Background(), in, from, from.AddDate(0, 0, 14))
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Upcoming) != 1 || snap.Upcoming[0].Title != "Show" || snap.Upcoming[0].Detail != "S01E02 · Pilot" {
		t.Fatalf("upcoming: %+v", snap.Upcoming)
	}
	q := snap.Queue
	if len(q) != 2 || q[0].Progress != 75 || q[0].TimeLeft != 26*time.Hour+3*time.Minute+4*time.Second || q[1].Error != "No files found" {
		t.Fatalf("queue: %+v", q)
	}
	if snap.Missing != 7 || len(snap.Health) != 1 || snap.Health[0].Severity != domain.SeverityWarning {
		t.Fatalf("health and missing: %+v", snap)
	}
	in.APIKey = "bad"
	if err := s.Check(context.Background(), in); err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("expected key error, got %v", err)
	}
}

func TestRadarrAndProwlarr(t *testing.T) {
	srv := arrServer(t, "Radarr")
	defer srv.Close()
	from := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	r := NewRadarr(srv.Client())
	snap, err := r.Snapshot(context.Background(), domain.Instance{App: domain.AppRadarr, Name: "Movies", URL: srv.URL, APIKey: "key"}, from, from.AddDate(0, 0, 14))
	if err != nil || len(snap.Upcoming) != 1 || snap.Upcoming[0].Date.Day() != 20 || snap.Upcoming[0].Detail != "2026" {
		t.Fatalf("the release date inside the range is used: %+v %v", snap.Upcoming, err)
	}
	if err := NewSonarr(srv.Client()).Check(context.Background(), domain.Instance{App: domain.AppSonarr, URL: srv.URL, APIKey: "key"}); err == nil {
		t.Fatal("a Radarr URL saved as Sonarr is reported")
	}

	prow := arrServer(t, "Prowlarr")
	defer prow.Close()
	p := NewProwlarr(prow.Client())
	in := domain.Instance{App: domain.AppProwlarr, Name: "Indexers", URL: prow.URL, APIKey: "key"}
	if err := p.Check(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if snap, err := p.Snapshot(context.Background(), in, from, from); err != nil || len(snap.Health) != 1 || snap.Health[0].Instance != "Indexers" {
		t.Fatalf("prowlarr: %+v %v", snap, err)
	}
}

func TestParseTimeLeft(t *testing.T) {
	for in, want := range map[string]time.Duration{"00:01:30": 90 * time.Second, "2.00:00:00": 48 * time.Hour, "": 0, "x": 0} {
		if got := parseTimeLeft(in); got != want {
			t.Errorf("parseTimeLeft(%q) = %v, want %v", in, got, want)
		}
	}
}
