package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/arr/domain"
	"trisse-home/internal/modules/arr/usecase"
)

type fakeArr struct {
	usecase.Arr
	added usecase.InstanceInput
}

func (f *fakeArr) Instances(context.Context) ([]usecase.InstanceView, error) {
	return []usecase.InstanceView{{ID: "1", App: "sonarr", Name: "Sonarr", URL: "http://s"}}, nil
}

func (f *fakeArr) AddInstance(_ context.Context, in usecase.InstanceInput) (usecase.InstanceView, error) {
	f.added = in
	return usecase.InstanceView{ID: "2", App: in.App, Name: in.Name, URL: in.URL}, nil
}

func (f *fakeArr) Overview(context.Context) (domain.Overview, error) {
	return domain.Overview{Queue: []domain.QueueItem{{Title: "x", Progress: 33.3, TimeLeft: 90 * time.Second}}}, nil
}

func TestHandlers(t *testing.T) {
	fake := &fakeArr{}
	h := NewHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/arr/instances", h.Instances)
	mux.HandleFunc("POST /api/arr/instances", h.Add)
	mux.HandleFunc("GET /api/arr/overview", h.Overview)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arr/instances", nil))
	if strings.Contains(strings.ToLower(rec.Body.String()), "apikey") {
		t.Fatalf("API keys are never returned: %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arr/instances", strings.NewReader(`{"app":"radarr","url":"http://r","apiKey":"k"}`)))
	if rec.Code != http.StatusCreated || fake.added.APIKey != "k" || fake.added.App != "radarr" {
		t.Fatalf("add: %d %+v", rec.Code, fake.added)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arr/overview", nil))
	body := rec.Body.String()
	if !strings.Contains(body, `"progress":33`) || !strings.Contains(body, `"timeLeftSeconds":90`) || !strings.Contains(body, `"errors":[]`) {
		t.Fatalf("overview: %s", body)
	}
}
