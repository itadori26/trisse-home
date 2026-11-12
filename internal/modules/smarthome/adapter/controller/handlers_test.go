package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/smarthome/domain"
	"trisse-home/internal/modules/smarthome/usecase"
)

type fakeSmartHome struct {
	usecase.SmartHome
	ids   []string
	acted string
}

func (f *fakeSmartHome) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, URL: "http://ha", HasToken: true}, nil
}

func (f *fakeSmartHome) States(_ context.Context, ids []string) ([]usecase.EntityView, error) {
	f.ids = ids
	return []usecase.EntityView{{Entity: domain.Entity{ID: "switch.coffee", Name: "Coffee", State: "on"}, Action: domain.ActionToggle}}, nil
}

func (f *fakeSmartHome) Act(_ context.Context, id string) error {
	if id != "switch.coffee" {
		return domain.ErrActionNotAllowed
	}
	f.acted = id
	return nil
}

func TestHandlers(t *testing.T) {
	fake := &fakeSmartHome{}
	h := NewHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/smarthome/settings", h.Settings)
	mux.HandleFunc("GET /api/smarthome/states", h.States)
	mux.HandleFunc("POST /api/smarthome/entities/{id}/action", h.Act)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/smarthome/settings", nil))
	if strings.Contains(rec.Body.String(), `"token"`) || !strings.Contains(rec.Body.String(), `"hasToken":true`) {
		t.Fatalf("settings must not expose the token: %s", rec.Body)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/smarthome/states?ids=switch.coffee,+sensor.power,,", nil))
	if rec.Code != http.StatusOK || len(fake.ids) != 2 || fake.ids[1] != "sensor.power" || !strings.Contains(rec.Body.String(), `"action":"toggle"`) {
		t.Fatalf("states: %d %v %s", rec.Code, fake.ids, rec.Body)
	}

	for path, want := range map[string]int{
		"/api/smarthome/entities/switch.coffee/action": http.StatusNoContent,
		"/api/smarthome/entities/switch.heater/action": http.StatusForbidden,
	} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, nil))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", path, rec.Code, want)
		}
	}
}
