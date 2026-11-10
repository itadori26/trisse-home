package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/dashboard/usecase"
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
	h := NewHandlers(nil, nil, fw)
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
