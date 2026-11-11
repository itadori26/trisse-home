package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
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
