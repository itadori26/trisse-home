package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/modules/docker/usecase"
)

type fakeUpdates struct {
	usecase.ImageUpdates
	saved usecase.UpdateSettingsInput
}

func (f *fakeUpdates) List(context.Context) ([]usecase.ImageUpdateView, error) {
	return []usecase.ImageUpdateView{{ImageUpdate: usecase.ImageUpdate{Ref: "nginx:1.27", Status: domain.UpdateAvailable, NewerTag: "1.28"}, Containers: []string{"web"}}}, nil
}

func (f *fakeUpdates) SaveSettings(_ context.Context, in usecase.UpdateSettingsInput) (usecase.UpdateSettingsView, error) {
	f.saved = in
	return usecase.UpdateSettingsView{Enabled: in.Enabled, Registries: []usecase.RegistryView{{Host: "ghcr.io", Username: "me", HasPassword: true}}}, nil
}

func TestUpdateHandlers(t *testing.T) {
	uc := &fakeUpdates{}
	h := NewUpdateHandlers(uc)
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/docker/updates", nil))
	if !strings.Contains(rec.Body.String(), `"status":"update_available","newerTag":"1.28"`) {
		t.Fatalf("%s", rec.Body)
	}
	rec = httptest.NewRecorder()
	body := `{"enabled":true,"semverTags":false,"registries":[{"host":"ghcr.io","username":"me","password":"pw"}]}`
	h.SaveSettings(rec, httptest.NewRequest(http.MethodPut, "/api/docker/updates/settings", strings.NewReader(body)))
	if rec.Code != http.StatusOK || uc.saved.Credentials[0].Password != "pw" || strings.Contains(rec.Body.String(), "pw\"") {
		t.Fatalf("%d %s %+v", rec.Code, rec.Body, uc.saved)
	}
}
