package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/modules/holidays/usecase"
)

type fakeUC struct {
	usecase.Holidays
	state domain.State
	n     int
}

func (f *fakeUC) Upcoming(_ context.Context, state domain.State, n int) (usecase.Overview, error) {
	f.state, f.n = state, n
	d := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	return usecase.Overview{State: state, Entries: []domain.Entry{{Kind: domain.KindPublic, Key: "christmas_day", Start: d, End: d, DaysUntil: 5}}}, nil
}

func TestUpcoming(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)
	rec := httptest.NewRecorder()
	h.Upcoming(rec, httptest.NewRequest(http.MethodGet, "/api/holidays?state=by&limit=5", nil))
	if rec.Code != http.StatusOK || uc.state != "BY" || uc.n != 5 ||
		!strings.Contains(rec.Body.String(), `"key":"christmas_day"`) || !strings.Contains(rec.Body.String(), `"start":"2026-12-25"`) ||
		!strings.Contains(rec.Body.String(), `"schoolUpdatedAt":null`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	h.Upcoming(rec, httptest.NewRequest(http.MethodGet, "/api/holidays?state=xx", nil))
	if uc.state != domain.DefaultState || uc.n != 8 {
		t.Fatalf("defaults: %s %d", uc.state, uc.n)
	}
}
