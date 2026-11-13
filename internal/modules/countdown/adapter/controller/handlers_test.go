package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/countdown/domain"
	"trisse-home/internal/modules/countdown/usecase"
)

type fakeUC struct {
	usecase.Countdowns
	grace int
}

func (f *fakeUC) List(_ context.Context, grace int) ([]domain.Countdown, error) {
	f.grace = grace
	c, _ := domain.New("1", domain.Input{Title: "Exam", Date: "2026-12-24", Time: "09:30"}, time.Time{})
	return []domain.Countdown{c}, nil
}

func (f *fakeUC) Create(_ context.Context, in domain.Input) (domain.Countdown, error) {
	return domain.New("1", in, time.Time{})
}

func (f *fakeUC) Update(_ context.Context, id string, in domain.Input) (domain.Countdown, error) {
	if id != "1" {
		return domain.Countdown{}, domain.ErrNotFound
	}
	return domain.New(id, in, time.Time{})
}

func TestCreate(t *testing.T) {
	for body, want := range map[string]int{
		`{"title":"Exam","date":"2026-12-24","time":"09:30"}`: http.StatusCreated,
		`{"title":"","date":"2026-12-24"}`:                    http.StatusUnprocessableEntity,
		`{"title":"Exam","date":"tomorrow"}`:                  http.StatusUnprocessableEntity,
	} {
		rec := httptest.NewRecorder()
		NewHandlers(&fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/countdowns", strings.NewReader(body)))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", body, rec.Code, want)
		}
	}
}

func TestUpdate(t *testing.T) {
	for id, want := range map[string]int{"1": http.StatusOK, "2": http.StatusNotFound} {
		req := httptest.NewRequest(http.MethodPut, "/api/countdowns/"+id, strings.NewReader(`{"title":"X","date":"2026-12-24"}`))
		req.SetPathValue("id", id)
		rec := httptest.NewRecorder()
		NewHandlers(&fakeUC{}).Update(rec, req)
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", id, rec.Code, want)
		}
	}
}

func TestListGrace(t *testing.T) {
	f := &fakeUC{}
	rec := httptest.NewRecorder()
	NewHandlers(f).List(rec, httptest.NewRequest(http.MethodGet, "/api/countdowns", nil))
	if rec.Code != http.StatusOK || f.grace != DefaultGraceDays || !strings.Contains(rec.Body.String(), `"time":"09:30"`) {
		t.Fatalf("%d %d %s", rec.Code, f.grace, rec.Body)
	}
	NewHandlers(f).List(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/countdowns?graceDays=7", nil))
	if f.grace != 7 {
		t.Fatalf("grace %d", f.grace)
	}
	rec = httptest.NewRecorder()
	NewHandlers(f).List(rec, httptest.NewRequest(http.MethodGet, "/api/countdowns?graceDays=x", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad grace: %d", rec.Code)
	}
}
