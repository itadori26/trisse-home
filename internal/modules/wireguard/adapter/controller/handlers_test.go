package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/wireguard/domain"
	"trisse-home/internal/modules/wireguard/usecase"
)

type fakeUC struct{ usecase.WireGuard }

func (fakeUC) List(context.Context) (usecase.ListView, error) {
	return usecase.ListView{}, domain.ErrNoExport
}

func TestListWithoutExport(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).List(rec, httptest.NewRequest(http.MethodGet, "/api/wireguard/peers", nil))
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "wireguard-export.sh") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
