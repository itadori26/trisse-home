package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/notes/domain"
	"trisse-home/internal/modules/notes/usecase"
)

type fakeSave struct{ gotID string }

func (f *fakeSave) Execute(_ context.Context, id string, in usecase.NoteInput) (usecase.NoteView, error) {
	f.gotID = id
	if id == "missing" {
		return usecase.NoteView{}, domain.ErrNoteNotFound
	}
	return usecase.NoteView{ID: "n1", Title: in.Title}, nil
}

func TestSaveHandlers(t *testing.T) {
	fs := &fakeSave{}
	h := NewHandlers(UseCases{Save: fs})
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/notes", h.Create)
	mux.HandleFunc("PUT /api/notes/{id}", h.Update)
	for _, tc := range []struct {
		method, path string
		want         int
	}{
		{http.MethodPost, "/api/notes", http.StatusCreated},
		{http.MethodPut, "/api/notes/n1", http.StatusOK},
		{http.MethodPut, "/api/notes/missing", http.StatusNotFound},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"title":"x","body":"","pinned":false}`)))
		if rec.Code != tc.want {
			t.Errorf("%s %s: status %d, want %d", tc.method, tc.path, rec.Code, tc.want)
		}
	}
}
