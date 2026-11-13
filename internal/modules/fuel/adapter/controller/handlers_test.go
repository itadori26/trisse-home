package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/fuel/domain"
	"trisse-home/internal/modules/fuel/usecase"
)

type fakeFuel struct {
	usecase.Fuel
	favs []domain.Favourite
}

func (f *fakeFuel) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, HasAPIKey: true, Fuel: "e5", Favourites: []domain.Favourite{}}, nil
}

func (f *fakeFuel) SetFavourites(_ context.Context, list []domain.Favourite) (usecase.SettingsView, error) {
	f.favs = list
	return usecase.SettingsView{Configured: true, Favourites: list}, nil
}

func (f *fakeFuel) Overview(context.Context) (domain.Overview, error) {
	return domain.Overview{Fuel: domain.FuelE5, Cheapest: []domain.Station{{ID: "a", Price: 1.799, DistanceKm: 1.26}, {ID: "b"}}}, nil
}

func TestHandlers(t *testing.T) {
	fake := &fakeFuel{}
	h := NewHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/fuel/settings", h.Settings)
	mux.HandleFunc("PUT /api/fuel/favourites", h.Favourites)
	mux.HandleFunc("GET /api/fuel/overview", h.Overview)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/fuel/settings", nil))
	if strings.Contains(rec.Body.String(), `"apiKey"`) || !strings.Contains(rec.Body.String(), `"hasApiKey":true`) {
		t.Fatalf("settings must not expose the key: %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/fuel/favourites", strings.NewReader(`[{"id":"s1","name":"Aral"}]`)))
	if rec.Code != http.StatusOK || len(fake.favs) != 1 || fake.favs[0].Name != "Aral" {
		t.Fatalf("favourites: %d %+v", rec.Code, fake.favs)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/fuel/overview", nil))
	body := rec.Body.String()
	if !strings.Contains(body, `"price":1.799`) || !strings.Contains(body, `"distanceKm":1.3`) || !strings.Contains(body, `"price":null`) {
		t.Fatalf("overview: %s", body)
	}
}
