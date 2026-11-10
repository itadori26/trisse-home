package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/modules/rss/usecase"
)

type fakeReader struct {
	usecase.Reader
	got usecase.ItemFilter
}

func (f *fakeReader) Items(_ context.Context, flt usecase.ItemFilter) ([]usecase.ItemView, error) {
	f.got = flt
	return nil, nil
}

func TestItemsQuery(t *testing.T) {
	fr := &fakeReader{}
	rec := httptest.NewRecorder()
	NewHandlers(fr).Items(rec, httptest.NewRequest(http.MethodGet, "/api/rss/items?feed=f1&unread=true&limit=7", nil))
	if rec.Code != http.StatusOK || fr.got != (usecase.ItemFilter{FeedID: "f1", UnreadOnly: true, Limit: 7}) {
		t.Fatalf("status %d filter %+v", rec.Code, fr.got)
	}
	if rec.Body.String() != "[]\n" {
		t.Fatalf("expected empty array, got %q", rec.Body.String())
	}
}
