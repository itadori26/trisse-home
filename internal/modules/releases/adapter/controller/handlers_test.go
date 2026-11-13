package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/modules/releases/usecase"
)

type fakeReleases struct {
	usecase.Releases
	weeks    int
	followed []string
}

func (f *fakeReleases) List(_ context.Context, weeks int) (usecase.ReleasesView, error) {
	f.weeks = weeks
	return usecase.ReleasesView{Releases: []domain.Release{{ID: "r", Title: "LP", Date: time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)}}}, nil
}

func (f *fakeReleases) Follow(_ context.Context, id, name string) error {
	if id == "bad" {
		return domain.ErrTooManyArtists
	}
	f.followed = append(f.followed, name)
	return nil
}

func TestHandlers(t *testing.T) {
	fake := &fakeReleases{}
	h := NewHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/releases", h.List)
	mux.HandleFunc("POST /api/releases/artists", h.Follow)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/releases?weeks=8", nil))
	if fake.weeks != 8 || !strings.Contains(rec.Body.String(), `"date":"2026-11-07"`) {
		t.Fatalf("list: %d %s", fake.weeks, rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/releases/artists", strings.NewReader(`{"id":"a","name":"Björk"}`)))
	if rec.Code != http.StatusNoContent || fake.followed[0] != "Björk" {
		t.Fatalf("follow: %d %v", rec.Code, fake.followed)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/releases/artists", strings.NewReader(`{"id":"bad","name":"x"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("limit: %d", rec.Code)
	}
}
