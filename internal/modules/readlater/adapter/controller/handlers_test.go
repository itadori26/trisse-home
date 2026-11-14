package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/modules/readlater/usecase"
)

type fakeUC struct {
	usecase.ReadLater
	status domain.Status
}

func (f *fakeUC) List(_ context.Context, s domain.Status) ([]domain.Article, error) {
	f.status = s
	return []domain.Article{{ID: "a", Title: "T", Words: 440, Status: domain.StatusUnread}}, nil
}

func (f *fakeUC) Image(context.Context, string, int) (usecase.Image, error) {
	return usecase.Image{Bytes: []byte("png"), ContentType: "image/png"}, nil
}

func TestHandlers(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/readlater?status=archived", nil))
	if rec.Code != http.StatusOK || uc.status != domain.StatusArchived || !strings.Contains(rec.Body.String(), `"readingMinutes":2`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	h.List(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/readlater?status=bogus", nil))
	if uc.status != "" {
		t.Fatal("unknown status lists unread and read")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/readlater/a/images/0", nil)
	req.SetPathValue("id", "a")
	req.SetPathValue("n", "0")
	rec = httptest.NewRecorder()
	h.Image(rec, req)
	if rec.Header().Get("Content-Type") != "image/png" || rec.Header().Get("X-Content-Type-Options") != "nosniff" ||
		!strings.Contains(rec.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Fatalf("%v", rec.Header())
	}
}
