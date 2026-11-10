package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/usecase"
)

type fakeCalendar struct {
	usecase.Calendar
	days int
}

func (f *fakeCalendar) Upcoming(_ context.Context, days int) (usecase.AgendaResponse, error) {
	f.days = days
	d := time.Date(2026, 12, 24, 0, 0, 0, 0, time.UTC)
	return usecase.AgendaResponse{Events: []usecase.EventView{{Summary: "Holiday", Start: d, End: d.AddDate(0, 0, 1), AllDay: true}}}, nil
}

func TestEvents(t *testing.T) {
	fc := &fakeCalendar{}
	rec := httptest.NewRecorder()
	NewHandlers(fc).Events(rec, httptest.NewRequest(http.MethodGet, "/api/calendar/events?days=14", nil))
	if rec.Code != http.StatusOK || fc.days != 14 || !strings.Contains(rec.Body.String(), `"start":"2026-12-24"`) {
		t.Fatalf("unexpected %d %s", rec.Code, rec.Body)
	}
}
