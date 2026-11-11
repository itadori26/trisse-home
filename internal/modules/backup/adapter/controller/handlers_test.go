package controller

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/modules/backup/usecase"
)

type fakeUC struct{ usecase.Backups }

func (fakeUC) Open(_ context.Context, name string) (io.ReadCloser, usecase.FileView, error) {
	if !domain.ValidName(name) {
		return nil, usecase.FileView{}, domain.ErrBackupNotFound
	}
	return io.NopCloser(strings.NewReader("SQLite")), usecase.FileView{Name: name, Size: 6}, nil
}

func TestDownload(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/backups/{name}", NewHandlers(fakeUC{}).Download)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/backups/trisse-home-20261111-030000.db", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "SQLite" || !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("unexpected %d %q %v", rec.Code, rec.Body, rec.Header())
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/backups/evil.db", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("invalid name: %d", rec.Code)
	}
}
