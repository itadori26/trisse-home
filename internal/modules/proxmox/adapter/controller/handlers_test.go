package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/modules/proxmox/usecase"
)

type fakeProxmox struct {
	usecase.Proxmox
	acted []string
}

func (f *fakeProxmox) Settings(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{Configured: true, URL: "https://pve", TokenID: "a@pve!t", HasSecret: true}, nil
}

func (f *fakeProxmox) Status(context.Context) (usecase.StatusView, error) {
	return usecase.StatusView{Cluster: domain.Cluster{Guests: []domain.Guest{{ID: 100, Kind: domain.KindVM, CPU: 0.1234}, {ID: 101, Kind: domain.KindLXC}}},
		Actions: []int{100}}, nil
}

func (f *fakeProxmox) Act(_ context.Context, id int, action string) error {
	if id != 100 {
		return domain.ErrActionNotAllowed
	}
	f.acted = append(f.acted, action)
	return nil
}

func TestHandlers(t *testing.T) {
	fake := &fakeProxmox{}
	h := NewHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/proxmox/settings", h.Settings)
	mux.HandleFunc("GET /api/proxmox/status", h.Status)
	mux.HandleFunc("POST /api/proxmox/guests/{id}/{action}", h.Act)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/proxmox/settings", nil))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"secret"`) || !strings.Contains(rec.Body.String(), `"hasSecret":true`) {
		t.Fatalf("settings must not expose the secret: %d %s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/proxmox/status", nil))
	body := rec.Body.String()
	if !strings.Contains(body, `"id":100,`) || !strings.Contains(body, `"cpuPercent":12.3`) || strings.Count(body, `"actions":true`) != 1 {
		t.Fatalf("status: %s", body)
	}

	for path, want := range map[string]int{
		"/api/proxmox/guests/100/start": http.StatusNoContent,
		"/api/proxmox/guests/101/start": http.StatusForbidden,
		"/api/proxmox/guests/abc/start": http.StatusNotFound,
	} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, nil))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", path, rec.Code, want)
		}
	}
	if len(fake.acted) != 1 || fake.acted[0] != "start" {
		t.Fatalf("acted: %v", fake.acted)
	}
}
