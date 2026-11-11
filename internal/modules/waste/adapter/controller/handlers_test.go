package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/modules/waste/usecase"
)

type fakeUC struct{ usecase.Waste }

func (fakeUC) Upcoming(context.Context) (usecase.Overview, error) {
	return usecase.Overview{}, domain.ErrNotConfigured
}

func (fakeUC) Settings(context.Context) (domain.Settings, bool, error) {
	return domain.Settings{Rules: []domain.Rule{}}, false, nil
}

func TestHandlers(t *testing.T) {
	h := NewHandlers(fakeUC{})
	rec := httptest.NewRecorder()
	h.Upcoming(rec, httptest.NewRequest(http.MethodGet, "/api/waste/upcoming", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("upcoming: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.Settings(rec, httptest.NewRequest(http.MethodGet, "/api/waste/settings", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != `{"configured":false,"url":"","rules":[]}`+"\n" {
		t.Fatalf("settings: %d %s", rec.Code, rec.Body)
	}
}
