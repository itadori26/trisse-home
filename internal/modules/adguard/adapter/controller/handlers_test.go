package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/adguard/usecase"
)

type fakeAdGuard struct {
	usecase.DNSFilter
	provider string
}

func (f *fakeAdGuard) Settings(_ context.Context, provider string) (usecase.SettingsView, error) {
	f.provider = provider
	return usecase.SettingsView{Provider: "adguard", Configured: true, URL: "http://ag", Username: "admin", HasPassword: true}, nil
}

func (f *fakeAdGuard) Stats(_ context.Context, provider string) (usecase.StatsView, error) {
	f.provider = provider
	return usecase.StatsView{Queries: 10}, nil
}

func TestSettingsNeverExposesPassword(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(&fakeAdGuard{}).Settings(rec, httptest.NewRequest(http.MethodGet, "/api/adguard/settings", nil))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"password"`) || !strings.Contains(rec.Body.String(), `"hasPassword":true`) {
		t.Fatalf("unexpected %d %s", rec.Code, rec.Body)
	}
}

func TestProviderParameter(t *testing.T) {
	fake := &fakeAdGuard{}
	rec := httptest.NewRecorder()
	NewHandlers(fake).Stats(rec, httptest.NewRequest(http.MethodGet, "/api/adguard/stats?provider=pihole", nil))
	if rec.Code != http.StatusOK || fake.provider != "pihole" {
		t.Fatalf("provider not passed: %d %q", rec.Code, fake.provider)
	}
	// Pi-hole reports no processing time: null instead of a misleading 0.
	if !strings.Contains(rec.Body.String(), `"avgProcessingMs":null`) {
		t.Fatalf("expected null average: %s", rec.Body)
	}
}

func (f *fakeAdGuard) Pause(_ context.Context, provider string, minutes int) (usecase.StatsView, error) {
	f.provider = provider
	until := time.Date(2026, 11, 14, 8, 30, 0, 0, time.UTC)
	return usecase.StatsView{PausedUntil: &until, Queries: minutes}, nil
}

func TestPause(t *testing.T) {
	fake := &fakeAdGuard{}
	rec := httptest.NewRecorder()
	NewHandlers(fake).Pause(rec, httptest.NewRequest(http.MethodPost, "/api/adguard/pause?provider=pihole", strings.NewReader(`{"minutes":30}`)))
	if rec.Code != http.StatusOK || fake.provider != "pihole" || !strings.Contains(rec.Body.String(), `"pausedUntil":"2026-11-14T08:30:00Z"`) ||
		!strings.Contains(rec.Body.String(), `"queries":30`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
