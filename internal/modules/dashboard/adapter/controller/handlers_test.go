package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"trisse-home/internal/modules/dashboard/domain"
	"trisse-home/internal/modules/dashboard/usecase"
	"trisse-home/internal/shared/httpapi"
)

type fakeWidgets struct {
	usecase.ManageWidgets
	got usecase.AddWidgetRequest
}

func (f *fakeWidgets) Add(_ context.Context, req usecase.AddWidgetRequest) (usecase.WidgetResponse, error) {
	f.got = req
	return usecase.WidgetResponse{ID: "w1", TabID: req.TabID, Kind: req.Kind, W: req.W, H: req.H, Config: req.Config}, nil
}

func TestAddWidget(t *testing.T) {
	fw := &fakeWidgets{}
	h := NewHandlers(nil, nil, fw, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/dashboard/tabs/{id}/widgets", h.AddWidget)

	tests := []struct {
		name, body string
		want       int
	}{
		{"valid", `{"kind":"weather","x":0,"y":0,"w":3,"h":2,"config":{ "lat" : 53.1 }}`, http.StatusCreated},
		{"config must be object", `{"kind":"weather","w":3,"h":2,"config":[1,2]}`, http.StatusUnprocessableEntity},
		{"unknown field", `{"kind":"weather","foo":1}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/dashboard/tabs/t1/widgets", strings.NewReader(tt.body)))
			if rec.Code != tt.want {
				t.Fatalf("status %d, want %d: %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
	if fw.got.TabID != "t1" || fw.got.Config != `{"lat":53.1}` {
		t.Fatalf("unexpected request %+v", fw.got)
	}
}

type fakeTabs struct {
	usecase.ManageTabs
	id    string
	rules []domain.RuleInput
}

func (f *fakeTabs) SetRules(_ context.Context, id string, in []domain.RuleInput) (usecase.TabResponse, error) {
	if _, err := domain.NewTabRules(in); err != nil {
		return usecase.TabResponse{}, err
	}
	f.id, f.rules = id, in
	return usecase.TabResponse{ID: id, Rules: []usecase.RuleView{{Days: []int{1}, From: "07:00", To: "09:00"}}}, nil
}

func TestSetRules(t *testing.T) {
	ft := &fakeTabs{}
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/dashboard/tabs/{id}/rules", NewHandlers(nil, ft, nil, nil).SetRules)
	tests := []struct {
		name, body string
		want       int
	}{
		{"valid", `{"rules":[{"days":[1],"from":"07:00","to":"09:00","networks":["192.168.178.0/24"]}]}`, http.StatusOK},
		{"invalid network", `{"rules":[{"networks":["lan"]}]}`, http.StatusUnprocessableEntity},
		{"unknown field", `{"rules":[],"x":1}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/dashboard/tabs/t1/rules", strings.NewReader(tt.body)))
			if rec.Code != tt.want {
				t.Fatalf("status %d, want %d: %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
	if ft.id != "t1" || len(ft.rules) != 1 || ft.rules[0].Networks[0] != "192.168.178.0/24" || ft.rules[0].From != "07:00" {
		t.Fatalf("unexpected request %q %+v", ft.id, ft.rules)
	}
}

type fakeGet struct {
	usecase.GetDashboard
	client netip.Addr
}

func (f *fakeGet) ExecuteFor(_ context.Context, client netip.Addr) (usecase.DashboardResponse, error) {
	f.client = client
	return usecase.DashboardResponse{DefaultTabID: "eve"}, nil
}

func TestGetPassesClientAddress(t *testing.T) {
	fg := &fakeGet{}
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req = req.WithContext(httpapi.WithClientIP(req.Context(), netip.MustParseAddr("10.13.13.2")))
	rec := httptest.NewRecorder()
	NewHandlers(fg, nil, nil, nil).Get(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"defaultTabId":"eve"`) {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if fg.client.String() != "10.13.13.2" {
		t.Fatalf("client %v", fg.client)
	}
}
