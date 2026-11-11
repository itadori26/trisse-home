package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/modules/docker/usecase"
)

type fakeActions struct{ usecase.Actions }

func (fakeActions) Restart(_ context.Context, name string) error {
	if name != "homarr" {
		return domain.ErrActionNotAllowed
	}
	return nil
}

func TestRestartHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/docker/containers/{name}/restart", NewActionHandlers(fakeActions{}).Restart)
	for name, want := range map[string]int{"homarr": http.StatusNoContent, "nextcloud-app": http.StatusForbidden} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/docker/containers/"+name+"/restart", nil))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", name, rec.Code, want)
		}
	}
}
