package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/focus/domain"
	"trisse-home/internal/modules/focus/usecase"
)

type fakeUC struct {
	usecase.Focus
	recorded int
}

var now = time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)

func (f *fakeUC) RecordSession(_ context.Context, minutes int, endedAt time.Time) (domain.Session, error) {
	s, err := domain.NewSession("1", minutes, endedAt, now)
	if err == nil {
		f.recorded++
	}
	return s, err
}

func (f *fakeUC) WeeklyStats(context.Context) ([]domain.Day, error) {
	return domain.Week([]domain.Session{{Minutes: 25, EndedAt: now}}, now), nil
}

func TestRecord(t *testing.T) {
	f := &fakeUC{}
	for body, want := range map[string]int{
		`{"minutes":25,"endedAt":"2026-11-12T09:59:00Z"}`: http.StatusCreated,
		`{"minutes":0,"endedAt":"2026-11-12T09:59:00Z"}`:  http.StatusUnprocessableEntity,
		`{"minutes":25,"endedAt":"yesterday"}`:            http.StatusBadRequest,
	} {
		rec := httptest.NewRecorder()
		NewHandlers(f).Record(rec, httptest.NewRequest(http.MethodPost, "/api/focus/sessions", strings.NewReader(body)))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d (%s)", body, rec.Code, want, rec.Body)
		}
	}
	if f.recorded != 1 {
		t.Fatalf("recorded %d", f.recorded)
	}
}

func TestStats(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(&fakeUC{}).Stats(rec, httptest.NewRequest(http.MethodGet, "/api/focus/stats", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `{"date":"2026-11-12","sessions":1,"minutes":25}`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
