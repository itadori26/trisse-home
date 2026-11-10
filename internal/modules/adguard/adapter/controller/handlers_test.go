package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/adguard/usecase"
)

type fakeAdGuard struct{ usecase.AdGuard }

func (fakeAdGuard) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, URL: "http://ag", Username: "admin", HasPassword: true}, nil
}

func TestSettingsNeverExposesPassword(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeAdGuard{}).Settings(rec, httptest.NewRequest(http.MethodGet, "/api/adguard/settings", nil))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"password"`) || !strings.Contains(rec.Body.String(), `"hasPassword":true`) {
		t.Fatalf("unexpected %d %s", rec.Code, rec.Body)
	}
}
