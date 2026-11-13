package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/modules/weather/usecase"
)

type fakeHealthUC struct {
	region int
	city   string
	types  []domain.PollenType
}

func (f *fakeHealthUC) Options(context.Context) (usecase.HealthOptions, error) {
	return usecase.HealthOptions{Regions: []domain.PollenRegion{{ID: 32, Name: "NI"}}, Cities: []string{"Hannover"}}, nil
}

func (f *fakeHealthUC) Report(_ context.Context, region int, city string, types []domain.PollenType) (usecase.HealthReport, error) {
	f.region, f.city, f.types = region, city, types
	return usecase.HealthReport{
		Pollen: &domain.PollenForecast{Region: domain.PollenRegion{ID: 32}, Days: []domain.PollenDay{{Type: domain.Birch, Today: 5, Tomorrow: domain.NoPollenData}}},
		UV:     &domain.UVForecast{City: city, Today: 6, Tomorrow: 2},
	}, nil
}

func TestHealthHandlers(t *testing.T) {
	uc := &fakeHealthUC{}
	h := NewHealthHandlers(uc)
	rec := httptest.NewRecorder()
	h.Report(rec, httptest.NewRequest(http.MethodGet, "/api/weather/health?region=32&city=Hannover&types=birch,Graeser,nope", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || uc.region != 32 || uc.city != "Hannover" || len(uc.types) != 2 ||
		!strings.Contains(body, `"today":"2-3","tomorrow":"","todayBurden":"high","tomorrowBurden":"unknown"`) ||
		!strings.Contains(body, `"todayLevel":"high","tomorrowLevel":"low"`) {
		t.Fatalf("%d %s %+v", rec.Code, body, uc)
	}
	rec = httptest.NewRecorder()
	h.Options(rec, httptest.NewRequest(http.MethodGet, "/api/weather/health/options", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"cities":["Hannover"]`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
