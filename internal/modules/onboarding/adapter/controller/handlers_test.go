package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/onboarding/usecase"
)

type fakeUC struct{ completed, restarted bool }

func (f *fakeUC) Status(context.Context) (usecase.StatusResponse, error) {
	return usecase.StatusResponse{}, nil
}
func (f *fakeUC) Complete(context.Context) (usecase.StatusResponse, error) {
	f.completed = true
	return usecase.StatusResponse{Completed: true}, nil
}
func (f *fakeUC) Restart(context.Context) (usecase.StatusResponse, error) {
	f.restarted = true
	return usecase.StatusResponse{}, nil
}

func TestUpdate(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)
	for body, want := range map[string]string{`{"completed":true}`: `"completed":true`, `{"completed":false}`: `"completed":false`} {
		rec := httptest.NewRecorder()
		h.Update(rec, httptest.NewRequest(http.MethodPut, "/api/onboarding", strings.NewReader(body)))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("%s: %d %s", body, rec.Code, rec.Body)
		}
	}
	if !uc.completed || !uc.restarted {
		t.Fatal("use cases not called")
	}
}
