package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/modules/weather/usecase"
)

type fakeWeather struct{ usecase.Weather }

func (fakeWeather) GetForecast(_ context.Context, req usecase.ForecastRequest) (usecase.ForecastResponse, error) {
	return usecase.ForecastResponse{Lat: req.Lat, Lon: req.Lon, Temperature: 5}, nil
}

func TestForecastHandler(t *testing.T) {
	h := NewHandlers(fakeWeather{})
	for url, want := range map[string]int{
		"/api/weather?lat=53.1&lon=8.2": http.StatusOK,
		"/api/weather?lat=abc&lon=8.2":  http.StatusUnprocessableEntity,
		"/api/weather":                  http.StatusUnprocessableEntity,
	} {
		rec := httptest.NewRecorder()
		h.Forecast(rec, httptest.NewRequest(http.MethodGet, url, nil))
		if rec.Code != want {
			t.Errorf("%s: status %d, want %d", url, rec.Code, want)
		}
	}
}
