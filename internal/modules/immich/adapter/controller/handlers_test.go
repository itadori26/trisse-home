package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/immich/domain"
	"trisse-home/internal/modules/immich/usecase"
)

type fakeImmich struct{ usecase.Immich }

func (fakeImmich) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, URL: "http://immich", HasAPIKey: true}, nil
}

func (fakeImmich) Overview(context.Context) (domain.Overview, error) {
	return domain.Overview{Stats: domain.LibraryStats{Photos: 3, DiskUsed: 1, DiskTotal: 3}}, nil
}

func (fakeImmich) Thumbnail(_ context.Context, id string) (domain.Thumbnail, error) {
	if id != "ok" {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	return domain.Thumbnail{ContentType: "image/webp", Data: []byte("webp")}, nil
}

func TestHandlers(t *testing.T) {
	h := NewHandlers(fakeImmich{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/immich/settings", h.Settings)
	mux.HandleFunc("GET /api/immich/overview", h.Overview)
	mux.HandleFunc("GET /api/immich/thumbnails/{id}", h.Thumbnail)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/immich/settings", nil))
	if strings.Contains(rec.Body.String(), `"apiKey"`) || !strings.Contains(rec.Body.String(), `"hasApiKey":true`) {
		t.Fatalf("settings must not expose the key: %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/immich/overview", nil))
	if !strings.Contains(rec.Body.String(), `"diskPercent":33.3`) || !strings.Contains(rec.Body.String(), `"memories":[]`) {
		t.Fatalf("overview: %s", rec.Body)
	}
	for id, want := range map[string]int{"ok": http.StatusOK, "nope": http.StatusNotFound} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/immich/thumbnails/"+id, nil))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", id, rec.Code, want)
		}
	}
}
