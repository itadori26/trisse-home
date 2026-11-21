package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/chores/domain"
	"waylight/internal/modules/chores/usecase"
	"waylight/internal/shared/kernel"
)

type fakeUC struct{ usecase.Chores }

func (fakeUC) Create(_ context.Context, in domain.Input, visibility string) (usecase.ChoreView, error) {
	c, err := domain.New("1", in, kernel.Access{Owner: "anna", Visibility: kernel.Visibility(visibility)}, time.Time{})
	return usecase.ChoreView{Chore: c, Mine: true, DueNow: true}, err
}

func (fakeUC) Done(_ context.Context, id string) (usecase.ChoreView, error) {
	if id != "1" {
		return usecase.ChoreView{}, domain.ErrNotFound
	}
	return usecase.ChoreView{Chore: domain.Chore{ID: "1", Title: "Dishes", LastDoneBy: "Anna", LastDoneAt: time.Unix(1, 0)}}, nil
}

func TestCreate(t *testing.T) {
	rec := httptest.NewRecorder()
	body := `{"title":"Dishes","everyDays":1,"due":"2026-11-23","assignees":["Anna","Ben"],"visibility":"everyone"}`
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/chores", strings.NewReader(body)))
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusCreated || got["assignee"] != "Anna" || got["dueNow"] != true || got["lastDoneAt"] != nil {
		t.Fatalf("%d %v", rec.Code, got)
	}
	rec = httptest.NewRecorder()
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/chores", strings.NewReader(`{"title":"x","everyDays":0,"due":"2026-11-23"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("every 0 days: %d", rec.Code)
	}
}

func TestDone(t *testing.T) {
	for id, want := range map[string]int{"1": http.StatusOK, "2": http.StatusNotFound} {
		req := httptest.NewRequest(http.MethodPost, "/api/chores/"+id+"/done", nil)
		req.SetPathValue("id", id)
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{}).Done(rec, req)
		if rec.Code != want {
			t.Errorf("%s: %d", id, rec.Code)
		}
	}
}
