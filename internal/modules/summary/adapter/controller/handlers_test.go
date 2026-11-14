package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/summary/usecase"
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
