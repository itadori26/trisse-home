package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/services/adapter/presenter"
	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/modules/services/usecase"
)

type fakeDiscovery struct {
	usecase.Discovery
	saved usecase.DiscoveryInput
}

func (f *fakeDiscovery) Save(_ context.Context, in usecase.DiscoveryInput) (usecase.DiscoveryView, error) {
	f.saved = in
	s := domain.DiscoverySettings{Enabled: in.Enabled, IntervalMinutes: in.IntervalMinutes, Rule: domain.MergeRule(in.Rule)}
	if err := s.Validate(); err != nil {
		return usecase.DiscoveryView{}, err
	}
	return usecase.DiscoveryView{Available: true, Settings: s}, nil
}

func (f *fakeDiscovery) Sync(context.Context) (usecase.DiscoveryView, error) {
	return usecase.DiscoveryView{}, domain.ErrDiscoveryUnavailable
}

func (f *fakeDiscovery) Get(context.Context) (usecase.DiscoveryView, error) {
	return usecase.DiscoveryView{Available: true, Settings: domain.DefaultDiscoverySettings(),
		Last: usecase.SyncResult{At: time.Unix(10, 0).UTC(), Created: 2}}, nil
}

func TestDiscoveryHandlers(t *testing.T) {
	fake := &fakeDiscovery{}
	h := NewDiscoveryHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/services/discovery", h.Get)
	mux.HandleFunc("PUT /api/services/discovery", h.Save)
	mux.HandleFunc("POST /api/services/discovery/sync", h.Sync)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/services/discovery", nil))
	var dto presenter.DiscoveryDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("get: %d %v", rec.Code, err)
	}
	if dto.Rule != "labels" || dto.LastSync == nil || dto.LastSync.Created != 2 {
		t.Fatalf("get: %+v", dto)
	}

	rec = httptest.NewRecorder()
	body := `{"enabled":true,"intervalMinutes":15,"rule":"manual"}`
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/services/discovery", strings.NewReader(body)))
	if rec.Code != http.StatusOK || fake.saved != (usecase.DiscoveryInput{Enabled: true, IntervalMinutes: 15, Rule: "manual"}) {
		t.Fatalf("save: %d %+v", rec.Code, fake.saved)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/services/discovery", strings.NewReader(`{"intervalMinutes":0,"rule":"labels"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid save: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/services/discovery/sync", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sync without docker: %d", rec.Code)
	}
}
