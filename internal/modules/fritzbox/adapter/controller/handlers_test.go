package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/fritzbox/domain"
	"trisse-home/internal/modules/fritzbox/usecase"
)

type fakeUC struct{ usecase.FritzBox }

func (fakeUC) Status(context.Context) (usecase.StatusView, error) {
	return usecase.StatusView{}, domain.ErrNotConfigured
}

func TestStatusUnconfigured(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Status(rec, httptest.NewRequest(http.MethodGet, "/api/fritzbox/status", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}

type callsUC struct {
	usecase.FritzBox
	got domain.CallFilter
}

func (c *callsUC) Calls(_ context.Context, f domain.CallFilter) (usecase.CallsView, error) {
	c.got = f
	return usecase.CallsView{Calls: []domain.Call{{Type: domain.CallMissed, Number: "0441 123", Name: "Oma"}}}, nil
}

func TestCalls(t *testing.T) {
	uc := &callsUC{}
	h := NewHandlers(uc)
	rec := httptest.NewRecorder()
	h.Calls(rec, httptest.NewRequest(http.MethodGet, "/api/fritzbox/calls?filter=missed&limit=20", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"type":"missed"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if uc.got.Kind() != domain.CallsMissed || uc.got.Limit() != 20 {
		t.Fatalf("filter not passed: %+v", uc.got)
	}
	for _, q := range []string{"?filter=rejected", "?limit=3", "?limit=lots"} {
		rec := httptest.NewRecorder()
		h.Calls(rec, httptest.NewRequest(http.MethodGet, "/api/fritzbox/calls"+q, nil))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: %d", q, rec.Code)
		}
	}
}
