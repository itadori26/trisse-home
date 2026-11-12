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
	event   usecase.PingEvent
	message string
}

func (f *fakeUC) Ping(_ context.Context, token string, event usecase.PingEvent, message string) error {
	if token != "good" {
		return domain.ErrSourceNotFound
	}
	f.event, f.message = event, message
	return nil
}

func TestPing(t *testing.T) {
	uc := &fakeUC{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/backup-status/ping/{token}", NewHandlers(uc).Ping)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/backup-status/ping/good?status=fail", strings.NewReader("disk full")))
	if rec.Code != http.StatusNoContent || uc.event != usecase.PingFail || uc.message != "disk full" {
		t.Fatalf("unexpected %d %+v", rec.Code, uc)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/backup-status/ping/bad", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("bad token: %d", rec.Code)
	}
}

func TestPingEvents(t *testing.T) {
	uc := &fakeUC{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/checks/ping/{token}/{event}", NewHandlers(uc).Ping)
	for event, want := range map[string]usecase.PingEvent{"start": usecase.PingStart, "fail": usecase.PingFail} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/checks/ping/good/"+event, nil))
		if rec.Code != http.StatusNoContent || uc.event != want {
			t.Errorf("%s: %d %s", event, rec.Code, uc.event)
		}
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/checks/ping/good/explode", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown event: %d", rec.Code)
	}
}
