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
