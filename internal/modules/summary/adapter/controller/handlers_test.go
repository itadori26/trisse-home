package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/summary/usecase"
	"trisse-home/internal/shared/kernel"
)

type fakeUC struct {
	usecase.Summary
	got usecase.Request
}

func (f *fakeUC) Summarize(_ context.Context, r usecase.Request) (usecase.Result, error) {
	f.got = r
	return usecase.Result{Text: "- kurz", Model: "m", Truncated: true}, nil
}

func TestSummarize(t *testing.T) {
	uc := &fakeUC{}
	rec := httptest.NewRecorder()
	NewHandlers(uc).Summarize(rec, httptest.NewRequest(http.MethodPost, "/api/summary", strings.NewReader(`{"url":"https://x.example/a","language":"de"}`)))
	if rec.Code != http.StatusOK || uc.got.URL != "https://x.example/a" || uc.got.Language != "de" ||
		!strings.Contains(rec.Body.String(), `"summary":"- kurz"`) || !strings.Contains(rec.Body.String(), `"truncated":true`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func (f *fakeUC) Ask(_ context.Context, q, c, lang string, onChunk func(string) error) error {
	if q == "?" {
		return kernel.NewValidationError("question", "Ask a question after the question mark.")
	}
	f.got = usecase.Request{Text: q + "|" + c, Language: lang}
	_ = onChunk("Jelly")
	return onChunk("fin")
}

func TestAsk(t *testing.T) {
	uc := &fakeUC{}
	rec := httptest.NewRecorder()
	NewHandlers(uc).Ask(rec, httptest.NewRequest(http.MethodPost, "/api/summary/ask", strings.NewReader(`{"question":"? warum","context":"ctx","language":"de"}`)))
	if rec.Code != http.StatusOK || rec.Body.String() != "Jellyfin" || rec.Header().Get("Content-Type") != "text/plain; charset=utf-8" || uc.got.Text != "? warum|ctx" {
		t.Fatalf("%d %q %v", rec.Code, rec.Body, rec.Header())
	}
	rec = httptest.NewRecorder()
	NewHandlers(uc).Ask(rec, httptest.NewRequest(http.MethodPost, "/api/summary/ask", strings.NewReader(`{"question":"?"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("errors before the first chunk are JSON errors: %d", rec.Code)
	}
}
