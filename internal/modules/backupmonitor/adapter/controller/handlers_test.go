package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/backupmonitor/domain"
	"trisse-home/internal/modules/backupmonitor/usecase"
)

type fakeUC struct {
	usecase.Monitor
	success bool
	message string
}

func (f *fakeUC) Ping(_ context.Context, token string, success bool, message string) error {
	if token != "good" {
		return domain.ErrSourceNotFound
	}
	f.success, f.message = success, message
	return nil
}

func TestPing(t *testing.T) {
	uc := &fakeUC{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/backup-status/ping/{token}", NewHandlers(uc).Ping)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/backup-status/ping/good?status=fail", strings.NewReader("disk full")))
	if rec.Code != http.StatusNoContent || uc.success || uc.message != "disk full" {
		t.Fatalf("unexpected %d %+v", rec.Code, uc)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/backup-status/ping/bad", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("bad token: %d", rec.Code)
	}
}
