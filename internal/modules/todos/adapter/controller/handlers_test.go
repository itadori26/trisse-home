package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/usecase"
)

type fakeCreate struct{ got usecase.TodoInput }

func (f *fakeCreate) Execute(_ context.Context, in usecase.TodoInput) (usecase.TodoView, error) {
	f.got = in
	return usecase.TodoView{ID: "1", Title: in.Title, Due: in.Due}, nil
}

func TestCreateParsesDueDate(t *testing.T) {
	fc := &fakeCreate{}
	h := NewHandlers(UseCases{Create: fc})
	rec := httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(`{"title":"x","due":"2026-12-24"}`)))
	if rec.Code != http.StatusCreated || !fc.got.Due.Equal(time.Date(2026, 12, 24, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("status %d due %v", rec.Code, fc.got.Due)
	}
	if !strings.Contains(rec.Body.String(), `"due":"2026-12-24"`) {
		t.Fatalf("unexpected body %s", rec.Body)
	}
	rec = httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(`{"title":"x","due":"24.12.2026"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid date accepted: %d", rec.Code)
	}
}

type fakeUpdate struct {
	got  usecase.TodoInput
	done bool
}

func (f *fakeUpdate) Execute(_ context.Context, _ string, in usecase.TodoInput, done bool) (usecase.TodoView, error) {
	f.got, f.done = in, done
	return usecase.TodoView{ID: "1", Title: in.Title, Repeat: usecase.RepeatView{Kind: "daily", Every: 2}}, nil
}

func TestUpdatePassesScheduleOnlyWhenSent(t *testing.T) {
	fu := &fakeUpdate{}
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/todos/{id}", NewHandlers(UseCases{Update: fu}).Update)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(
		`{"title":"x","due":"2026-12-24","done":true,"repeat":{"kind":"daily","every":2,"days":[]},"reminder":{"at":"08:00","daysBefore":1}}`)))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"repeat":{"kind":"daily","every":2,"days":[]}`) {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if fu.got.Repeat == nil || fu.got.Repeat.Every != 2 || fu.got.Reminder == nil || fu.got.Reminder.DaysBefore != 1 || !fu.done {
		t.Fatalf("input %+v", fu.got)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/todos/1", strings.NewReader(`{"title":"x","due":""}`)))
	if rec.Code != http.StatusOK || fu.got.Repeat != nil || fu.got.Reminder != nil {
		t.Fatalf("omitted schedule must stay nil: %d %+v", rec.Code, fu.got)
	}
}
