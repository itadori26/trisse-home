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

type fakeSearch struct{ got string }

func (f *fakeSearch) Execute(_ context.Context, q string) ([]domain.SearchHit, error) {
	f.got = q
	return []domain.SearchHit{{ID: "n1", Title: []domain.Segment{{Text: "<b>Büro</b>", Hit: true}}}}, nil
}

type fakeRevision struct{}

func (fakeRevision) Execute(_ context.Context, id, rev string) (usecase.RevisionDetail, error) {
	if rev != "r1" {
		return usecase.RevisionDetail{}, usecase.ErrRevisionNotFound
	}
	return usecase.RevisionDetail{RevisionView: usecase.RevisionView{ID: rev}, Body: "old",
		Diff: []domain.DiffLine{{Op: domain.DiffRemove, Text: "old"}, {Op: domain.DiffAdd, Text: "new"}}}, nil
}

type fakeRestore struct{ id, rev string }

func (f *fakeRestore) Execute(_ context.Context, id, rev string) (usecase.NoteView, error) {
	f.id, f.rev = id, rev
	return usecase.NoteView{ID: id, Body: "old"}, nil
}

func TestSearchAndHistoryHandlers(t *testing.T) {
	search, restore := &fakeSearch{}, &fakeRestore{}
	h := NewHandlers(UseCases{Search: search, Revision: fakeRevision{}, Restore: restore})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/notes/search", h.Search)
	mux.HandleFunc("GET /api/notes/{id}/revisions/{rev}", h.Revision)
	mux.HandleFunc("POST /api/notes/{id}/revisions/{rev}/restore", h.Restore)
	do := func(method, path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		return rec
	}

	rec := do(http.MethodGet, "/api/notes/search?q=b%C3%BCro")
	// Highlights are data, never markup: the title is JSON-escaped text.
	if rec.Code != http.StatusOK || search.got != "büro" ||
		strings.Contains(rec.Body.String(), "<b>") || !strings.Contains(rec.Body.String(), `"hit":true`) {
		t.Fatalf("search %d %s", rec.Code, rec.Body)
	}
	rec = do(http.MethodGet, "/api/notes/n1/revisions/r1")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"diff":[{"op":"-","text":"old"},{"op":"+","text":"new"}]`) {
		t.Fatalf("revision %d %s", rec.Code, rec.Body)
	}
	if rec := do(http.MethodGet, "/api/notes/n1/revisions/r2"); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown revision %d", rec.Code)
	}
	if rec := do(http.MethodPost, "/api/notes/n1/revisions/r1/restore"); rec.Code != http.StatusOK || restore.id != "n1" || restore.rev != "r1" {
		t.Fatalf("restore %d %+v", rec.Code, restore)
	}
}
