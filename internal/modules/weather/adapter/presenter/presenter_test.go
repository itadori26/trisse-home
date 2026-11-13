package presenter

import (
	"encoding/json"
	"strings"
	"testing"

	"trisse-home/internal/modules/weather/usecase"
)

func TestForecastWindDirection(t *testing.T) {
	tests := []struct {
		name string
		in   usecase.ForecastResponse
		want string
	}{
		{"unknown stays null", usecase.ForecastResponse{}, `"windDirection":null`},
		{"north is not unknown", usecase.ForecastResponse{HasWindDirection: true, WindDirection: 0}, `"windDirection":0`},
		{"rounds to whole degrees", usecase.ForecastResponse{HasWindDirection: true, WindDirection: 314.6}, `"windDirection":315`},
		{"359.6 wraps to north", usecase.ForecastResponse{HasWindDirection: true, WindDirection: 359.6}, `"windDirection":0`},
	}
	for _, tc := range tests {
		raw, err := json.Marshal(Forecast(tc.in))
		if err != nil || !strings.Contains(string(raw), tc.want) {
			t.Errorf("%s: want %s in %s (%v)", tc.name, tc.want, raw, err)
		}
	}
}
