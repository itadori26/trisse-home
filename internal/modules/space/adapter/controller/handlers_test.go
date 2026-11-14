package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/space/domain"
	"trisse-home/internal/modules/space/usecase"
)

type fakeUC struct{ usecase.Space }

func (fakeUC) Picture(context.Context) (domain.Picture, error) {
	return domain.Picture{Date: "2026-11-13", Title: "Pleiades", MediaType: "image", URL: "https://apod.nasa.gov/a.jpg"}, nil
}

func (fakeUC) PictureImage(context.Context) (usecase.Image, error) {
	return usecase.Image{Bytes: []byte("jpg"), ContentType: "image/jpeg"}, nil
}

func (fakeUC) Settings(context.Context) (domain.Settings, error) {
	return domain.Settings{Launches: true, NASAKey: "secret"}, nil
}

func TestHandlers(t *testing.T) {
	h := NewHandlers(fakeUC{})
	rec := httptest.NewRecorder()
	h.Picture(rec, httptest.NewRequest(http.MethodGet, "/api/space/apod", nil))
	if !strings.Contains(rec.Body.String(), `"image":"/api/space/apod/image?d=2026-11-13"`) || strings.Contains(rec.Body.String(), "apod.nasa.gov/a.jpg") {
		t.Fatalf("the browser must get the proxied URL only: %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	h.Image(rec, httptest.NewRequest(http.MethodGet, "/api/space/apod/image", nil))
	if rec.Header().Get("Content-Type") != "image/jpeg" || rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Body.String() != "jpg" {
		t.Fatalf("%v %s", rec.Header(), rec.Body)
	}
	rec = httptest.NewRecorder()
	h.Settings(rec, httptest.NewRequest(http.MethodGet, "/api/space/settings", nil))
	if strings.Contains(rec.Body.String(), "secret") || !strings.Contains(rec.Body.String(), `"hasKey":true`) {
		t.Fatalf("%s", rec.Body)
	}
}
