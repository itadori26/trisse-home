package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/modules/bookmarks/usecase"
)

type fakeList struct{ got usecase.Filter }

func (f *fakeList) Execute(_ context.Context, flt usecase.Filter) (usecase.ListResponse, error) {
	f.got = flt
	return usecase.ListResponse{}, nil
}

type fakeCreate struct{}

func (fakeCreate) Execute(_ context.Context, in usecase.Input) (usecase.BookmarkView, error) {
	if in.URL == "" {
		_, err := domain.NewBookmark("x", domain.BookmarkInput{}, testTime)
		return usecase.BookmarkView{}, err
	}
	return usecase.BookmarkView{ID: "1", URL: in.URL}, nil
}

func TestHandlers(t *testing.T) {
	list := &fakeList{}
	h := NewHandlers(UseCases{List: list, Create: fakeCreate{}})

	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/bookmarks?folder=f1&tag=go&q=doc", nil))
	if rec.Code != http.StatusOK || list.got != (usecase.Filter{FolderID: "f1", Tag: "go", Query: "doc"}) {
		t.Fatalf("filter not passed: %d %+v", rec.Code, list.got)
	}
	if !strings.Contains(rec.Body.String(), `"bookmarks":[]`) {
		t.Fatalf("expected empty arrays, got %s", rec.Body)
	}

	rec = httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/bookmarks", strings.NewReader(`{"url":"https://go.dev"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/api/bookmarks", strings.NewReader(`{"url":""}`)))
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), `"field":"url"`) {
		t.Fatalf("validation not mapped: %d %s", rec.Code, rec.Body)
	}
}

var testTime = time.Unix(0, 0)
