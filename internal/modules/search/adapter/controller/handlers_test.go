package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/shared/kernel"
)

type fakeSearcher struct{ usecase.Searcher }

func (fakeSearcher) Resolve(_ context.Context, q string) (string, error) {
	if q == "" {
		return "", kernel.NewValidationError("q", "empty")
	}
	return "https://duckduckgo.com/?q=" + q, nil
}

func TestSearchRedirects(t *testing.T) {
	h := NewHandlers(fakeSearcher{})
	rec := httptest.NewRecorder()
	h.Search(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=go", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://duckduckgo.com/?q=go" {
		t.Fatalf("unexpected response %d %s", rec.Code, rec.Header().Get("Location"))
	}
	rec = httptest.NewRecorder()
	h.Search(rec, httptest.NewRequest(http.MethodGet, "/api/search", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty query status %d", rec.Code)
	}
}

func TestOpenSearch(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/search/opensearch.xml", nil)
	req.Host = "192.168.1.2:8090"
	NewHandlers(fakeSearcher{}).OpenSearch(rec, req)
	if !strings.Contains(rec.Body.String(), `template="http://192.168.1.2:8090/api/search?q={searchTerms}"`) {
		t.Fatalf("unexpected body %s", rec.Body)
	}
}
