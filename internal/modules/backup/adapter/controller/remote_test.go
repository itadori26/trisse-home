package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/modules/backup/usecase"
)

type fakeOffsite struct {
	usecase.Offsite
	tested usecase.RemoteSettingsInput
}

func (f *fakeOffsite) Settings(context.Context) (usecase.RemoteSettingsView, error) {
	return usecase.RemoteSettingsView{Configured: true, URL: "https://c/dav/", Username: "me", HasPassword: true, Retention: 7}, nil
}

func (f *fakeOffsite) Test(_ context.Context, in usecase.RemoteSettingsInput) error {
	f.tested = in
	return nil
}

func (f *fakeOffsite) List(context.Context) ([]usecase.FileView, error) {
	return nil, domain.ErrRemoteNotConfigured
}

func TestRemoteHandlers(t *testing.T) {
	uc := &fakeOffsite{}
	h := NewRemoteHandlers(uc)
	rec := httptest.NewRecorder()
	h.Settings(rec, httptest.NewRequest(http.MethodGet, "/api/backups/remote/settings", nil))
	if strings.Contains(rec.Body.String(), `"password"`) || !strings.Contains(rec.Body.String(), `"hasPassword":true`) {
		t.Fatalf("%s", rec.Body)
	}
	rec = httptest.NewRecorder()
	h.Test(rec, httptest.NewRequest(http.MethodPost, "/api/backups/remote/test", strings.NewReader(`{"url":"https://c/dav/","username":"me","retention":5}`)))
	if rec.Code != http.StatusNoContent || uc.tested.Retention != 5 {
		t.Fatalf("%d %+v", rec.Code, uc.tested)
	}
	rec = httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/backups/remote", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unconfigured list: %d", rec.Code)
	}
}
