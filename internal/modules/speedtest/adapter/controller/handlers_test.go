package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/speedtest/domain"
	"trisse-home/internal/modules/speedtest/usecase"
)

type fakeUC struct {
	usecase.Speedtest
	hours int
}

func (f *fakeUC) Results(_ context.Context, hours int) (usecase.ResultsView, error) {
	f.hours = hours
	m := domain.Measurement{At: time.Unix(0, 0), DownloadBps: 123456789, UploadBps: 1e6, PingMs: 9.87}
	return usecase.ResultsView{Latest: &m, History: []domain.Measurement{m}}, nil
}

func (f *fakeUC) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, URL: "http://nas", HasToken: true}, nil
}

func TestResults(t *testing.T) {
	uc := &fakeUC{}
	rec := httptest.NewRecorder()
	NewHandlers(uc).Results(rec, httptest.NewRequest(http.MethodGet, "/api/speedtest/results?hours=24", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || uc.hours != 24 || !strings.Contains(body, `"downloadMbps":123.5`) || !strings.Contains(body, `"pingMs":9.9`) {
		t.Fatalf("%d %s", rec.Code, body)
	}
}

func TestSettingsNeverReturnsToken(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(&fakeUC{}).Settings(rec, httptest.NewRequest(http.MethodGet, "/api/speedtest/settings", nil))
	if strings.Contains(rec.Body.String(), `"token"`) || !strings.Contains(rec.Body.String(), `"hasToken":true`) {
		t.Fatalf("%s", rec.Body)
	}
}
