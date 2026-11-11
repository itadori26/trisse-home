package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/modules/transit/domain"
	"trisse-home/internal/modules/transit/usecase"
)

type fakeUC struct{ usecase.Transit }

func (fakeUC) Departures(_ context.Context, stop string, _ int) ([]domain.Departure, error) {
	if stop == "" {
		return nil, domain.ErrInvalidStop
	}
	return nil, nil
}

func TestDepartures(t *testing.T) {
	h := NewHandlers(fakeUC{})
	for q, want := range map[string]int{"": http.StatusUnprocessableEntity, "?stop=s1&n=3": http.StatusOK} {
		rec := httptest.NewRecorder()
		h.Departures(rec, httptest.NewRequest(http.MethodGet, "/api/transit/departures"+q, nil))
		if rec.Code != want {
			t.Errorf("%q: %d", q, rec.Code)
		}
	}
}
