package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/weather/domain"
)

const forecastJSON = `{"timezone":"Europe/Berlin",
 "current":{"temperature_2m":7.4,"apparent_temperature":4.1,"relative_humidity_2m":81,"wind_speed_10m":14.2,"wind_direction_10m":315,"weather_code":3,"is_day":1},
 "daily":{"time":["2026-11-10","2026-11-11"],"weather_code":[3,61],"temperature_2m_max":[9.1,8],"temperature_2m_min":[4.2,3.5],"precipitation_probability_max":[20,null]}}`

func TestOpenMeteo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/forecast":
			if !strings.Contains(r.URL.Query().Get("current"), "wind_direction_10m") {
				t.Errorf("the wind direction must be requested, got %q", r.URL.Query().Get("current"))
			}
			if r.URL.Query().Get("latitude") != "53.14" {
				t.Errorf("unexpected latitude %q", r.URL.Query().Get("latitude"))
			}
			_, _ = w.Write([]byte(forecastJSON))
		case "/search":
			_, _ = w.Write([]byte(`{"results":[{"name":"Oldenburg","admin1":"Lower Saxony","country":"Germany","latitude":53.14,"longitude":8.21}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	om := NewOpenMeteo(srv.Client(), srv.URL+"/forecast", srv.URL+"/search", "en")
	loc, _ := domain.NewLocation(53.1435, 8.2146)

	f, err := om.Forecast(context.Background(), loc)
	if err != nil {
		t.Fatal(err)
	}
	if f.Current.Temperature != 7.4 || !f.Current.IsDay || len(f.Days) != 2 || f.Days[1].PrecipitationProbability != 0 {
		t.Fatalf("unexpected forecast %+v", f)
	}
	if !f.Current.HasWindDirection || f.Current.WindDirection != 315 {
		t.Fatalf("wind direction %v (known: %v)", f.Current.WindDirection, f.Current.HasWindDirection)
	}
	places, err := om.SearchPlaces(context.Background(), "Oldenburg")
	if err != nil || len(places) != 1 || places[0].Region != "Lower Saxony" {
		t.Fatalf("unexpected places %+v %v", places, err)
	}

	bad := NewOpenMeteo(srv.Client(), srv.URL+"/missing", srv.URL+"/missing", "en")
	if _, err := bad.Forecast(context.Background(), loc); err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestForecastWithoutWindDirection(t *testing.T) {
	for name, raw := range map[string]string{
		"absent": `{"timezone":"UTC","current":{"temperature_2m":1,"wind_speed_10m":2,"weather_code":0,"is_day":1},"daily":{"time":[]}}`,
		"null":   `{"timezone":"UTC","current":{"temperature_2m":1,"wind_direction_10m":null,"weather_code":0,"is_day":1},"daily":{"time":[]}}`,
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(raw)) }))
		om := NewOpenMeteo(srv.Client(), srv.URL, srv.URL, "en")
		loc, _ := domain.NewLocation(1, 1)
		f, err := om.Forecast(context.Background(), loc)
		srv.Close()
		if err != nil || f.Current.HasWindDirection {
			t.Errorf("%s: an unknown direction must not read as north: %+v %v", name, f.Current, err)
		}
	}
}

func TestForecastNormalisesWindDirection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"timezone":"UTC","current":{"wind_direction_10m":-45,"weather_code":0,"is_day":1},"daily":{"time":[]}}`))
	}))
	defer srv.Close()
	om := NewOpenMeteo(srv.Client(), srv.URL, srv.URL, "en")
	loc, _ := domain.NewLocation(1, 1)
	f, err := om.Forecast(context.Background(), loc)
	if err != nil || f.Current.WindDirection != 315 {
		t.Fatalf("-45° must become 315°: %+v %v", f.Current, err)
	}
}
