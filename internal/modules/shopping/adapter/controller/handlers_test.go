package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/modules/shopping/usecase"
)

type fakeUC struct{ usecase.Shopping }

func (fakeUC) SetDone(_ context.Context, id string, _ bool) error {
	if id == "busy.ics" {
		return domain.ErrConflict
	}
	return nil
}

func TestSetDone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/shopping/items/{id}", NewHandlers(fakeUC{}).SetDone)
	for id, want := range map[string]int{"a.ics": http.StatusNoContent, "busy.ics": http.StatusConflict} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/shopping/items/"+id, strings.NewReader(`{"done":true}`)))
		if rec.Code != want {
			t.Errorf("%s: %d", id, rec.Code)
		}
	}
}
