package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/gifts/domain"
	"waylight/internal/modules/gifts/usecase"
)

type fakeUC struct{ usecase.Gifts }

func (fakeUC) Create(_ context.Context, in domain.Input) (domain.Idea, error) {
	return domain.New("1", "anna", in, time.Time{})
}

func TestCreate(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/gifts", strings.NewReader(`{"person":"Ben","title":"Book"}`)))
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusCreated || got["person"] != "Ben" || got["bought"] != false {
		t.Fatalf("%d %v", rec.Code, got)
	}
	if _, leaked := got["owner"]; leaked {
		t.Fatal("the owner is not part of the response")
	}
	rec = httptest.NewRecorder()
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/gifts", strings.NewReader(`{"person":"","title":"Book"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("no person: %d", rec.Code)
	}
}
