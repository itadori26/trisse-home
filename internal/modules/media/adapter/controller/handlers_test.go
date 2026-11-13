package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/media/domain"
	"trisse-home/internal/modules/media/usecase"
)

type fakeMedia struct{ usecase.Media }

func (fakeMedia) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, Provider: "plex", URL: "http://plex", HasToken: true}, nil
}

func (fakeMedia) Overview(context.Context) (domain.Overview, error) {
	return domain.Overview{Sessions: []domain.Session{{Title: "Film", Position: time.Minute, Duration: 3 * time.Minute}}}, nil
}

func (fakeMedia) Poster(_ context.Context, id string) (domain.Poster, error) {
	if id != "42" {
		return domain.Poster{}, domain.ErrPosterNotFound
	}
	return domain.Poster{ContentType: "image/jpeg", Data: []byte("jpeg")}, nil
}

func TestHandlers(t *testing.T) {
	h := NewHandlers(fakeMedia{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/media/settings", h.Settings)
	mux.HandleFunc("GET /api/media/overview", h.Overview)
	mux.HandleFunc("GET /api/media/posters/{id}", h.Poster)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/settings", nil))
	if strings.Contains(rec.Body.String(), `"token"`) || !strings.Contains(rec.Body.String(), `"hasToken":true`) {
		t.Fatalf("settings must not expose the token: %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/overview", nil))
	if !strings.Contains(rec.Body.String(), `"progress":33`) || !strings.Contains(rec.Body.String(), `"recent":[]`) {
		t.Fatalf("overview: %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/posters/42", nil))
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("poster: %d %v", rec.Code, rec.Header())
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/posters/7", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown poster: %d", rec.Code)
	}
}
