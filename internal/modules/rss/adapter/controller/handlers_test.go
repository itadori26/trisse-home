package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/rss/domain"
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

type opmlReader struct {
	usecase.Reader
	got []byte
}

func (f *opmlReader) ImportOPML(_ context.Context, data []byte) (usecase.ImportReport, error) {
	f.got = data
	return usecase.ImportReport{Added: 2, Duplicates: 1, Invalid: []domain.InvalidEntry{{Title: "Bad", URL: "x", Reason: "not an http(s) feed address"}}}, nil
}

func (f *opmlReader) ExportOPML(context.Context) ([]byte, error) { return []byte("<opml/>"), nil }

func TestOPMLHandlers(t *testing.T) {
	fr := &opmlReader{}
	h := NewHandlers(fr)
	rec := httptest.NewRecorder()
	h.ImportOPML(rec, httptest.NewRequest(http.MethodPost, "/api/rss/opml", strings.NewReader("<opml/>")))
	if rec.Code != http.StatusOK || string(fr.got) != "<opml/>" ||
		!strings.Contains(rec.Body.String(), `"added":2,"duplicates":1,"invalid":[{"title":"Bad"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	h.ImportOPML(rec, httptest.NewRequest(http.MethodPost, "/api/rss/opml", strings.NewReader(strings.Repeat("x", domain.MaxImportBytes+1))))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ExportOPML(rec, httptest.NewRequest(http.MethodGet, "/api/rss/opml", nil))
	if rec.Body.String() != "<opml/>" || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/x-opml") ||
		!strings.Contains(rec.Header().Get("Content-Disposition"), "trisse-home-feeds.opml") {
		t.Fatalf("%v %s", rec.Header(), rec.Body)
	}
}
