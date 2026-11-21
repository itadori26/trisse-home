package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/nextcloud/domain"
	"waylight/internal/modules/nextcloud/usecase"
)

type fakeUC struct{ usecase.Nextcloud }

func (fakeUC) Status(context.Context, string) (usecase.StatusView, error) {
	return usecase.StatusView{}, domain.ErrNotConfigured
}

func (fakeUC) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, URL: "https://c", Username: "anna", HasPassword: true}, nil
}

func TestHandlers(t *testing.T) {
	h := NewHandlers(fakeUC{})
	rec := httptest.NewRecorder()
	h.Status(rec, httptest.NewRequest(http.MethodGet, "/api/nextcloud/status", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unconfigured status: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.Settings(rec, httptest.NewRequest(http.MethodGet, "/api/nextcloud/settings", nil))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "password\"") {
		t.Fatalf("settings: %d %s", rec.Code, rec.Body)
	}
}
