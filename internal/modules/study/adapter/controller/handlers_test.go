package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/modules/study/usecase"
)

type fakeUC struct {
	usecase.Study
	got domain.ModuleInput
}

func (f *fakeUC) Update(_ context.Context, id string, in domain.ModuleInput) (domain.Module, error) {
	f.got = in
	return domain.Module{ID: id, Name: in.Name, ECTS: in.ECTS, Status: in.Status, Grade: in.Grade}, nil
}

func (f *fakeUC) Progress(context.Context) (domain.Progress, []domain.Module, error) {
	avg := 1.75
	ms := []domain.Module{{ID: "a", Name: "Analysis", ECTS: 9, Status: domain.StatusInProgress}, {ID: "b", Name: "DB", ECTS: 6, Status: domain.StatusPassed}}
	return domain.Progress{TotalECTS: 180, PassedECTS: 6, Average: &avg}, ms, nil
}

func TestHandlers(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)
	req := httptest.NewRequest(http.MethodPut, "/api/study/modules/a", strings.NewReader(`{"name":"DB","ects":7.5,"semester":3,"status":"passed","grade":1.3}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "a")
	rec := httptest.NewRecorder()
	h.Update(rec, req)
	if rec.Code != http.StatusOK || uc.got.ECTS != 7.5 || uc.got.Status != domain.StatusPassed || *uc.got.Grade != 1.3 {
		t.Fatalf("%d %+v", rec.Code, uc.got)
	}
	rec = httptest.NewRecorder()
	h.Progress(rec, httptest.NewRequest(http.MethodGet, "/api/study/progress", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"average":1.75`) || !strings.Contains(body, `"inProgress":[{"id":"a"`) || strings.Contains(body, `"name":"DB"`) {
		t.Fatalf("%d %s", rec.Code, body)
	}
}
