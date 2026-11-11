package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/transfer/usecase"
)

type fakeUC struct{ usecase.Transfer }

func TestImportRejectsGarbage(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Import(rec, httptest.NewRequest(http.MethodPost, "/api/transfer/import", strings.NewReader("<html>")))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got %d", rec.Code)
	}
	_ = context.Background()
}
