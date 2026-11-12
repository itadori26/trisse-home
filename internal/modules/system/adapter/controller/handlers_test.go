package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/modules/system/usecase"
)

type fakeMonitor struct{ usecase.Monitor }

func (fakeMonitor) Current(context.Context) (usecase.SnapshotView, error) {
	return usecase.SnapshotView{CPUPercent: 12.345, Memory: usecase.UsageView{Total: 100, Used: 25, Percent: 25}}, nil
}

type fakeForecasts struct{ usecase.DiskForecasts }

func (fakeForecasts) Forecasts(context.Context) (map[string]usecase.ForecastView, error) {
	return map[string]usecase.ForecastView{"System": {Status: domain.ForecastFilling, DaysUntilFull: 23.04}}, nil
}

func (fakeForecasts) SaveSettings(_ context.Context, s domain.ForecastSettings) (domain.ForecastSettings, error) {
	return s, s.Validate()
}

func TestSnapshotRoundsPercentages(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeMonitor{}, fakeForecasts{}).Snapshot(rec, httptest.NewRequest(http.MethodGet, "/api/system", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"cpuPercent":12.3`) || !strings.Contains(rec.Body.String(), `"disks":[]`) {
		t.Fatalf("unexpected %d %s", rec.Code, rec.Body)
	}
}

type diskMonitor struct{ usecase.Monitor }

func (diskMonitor) Current(context.Context) (usecase.SnapshotView, error) {
	return usecase.SnapshotView{Disks: []usecase.DiskView{{Label: "System"}, {Label: "Data"}}}, nil
}

func TestSnapshotWithForecast(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(diskMonitor{}, fakeForecasts{}).Snapshot(rec, httptest.NewRequest(http.MethodGet, "/api/system", nil))
	body := rec.Body.String()
	if !strings.Contains(body, `"status":"filling","daysUntilFull":23`) || !strings.Contains(body, `"label":"Data","usage":{"total":0,"used":0,"percent":0},"forecast":null`) {
		t.Fatalf("%s", body)
	}
}

func TestSaveForecastSettingsValidates(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/system/forecast-settings", strings.NewReader(`{"windowDays":1,"alertDays":14}`))
	NewHandlers(diskMonitor{}, fakeForecasts{}).SaveForecastSettings(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
