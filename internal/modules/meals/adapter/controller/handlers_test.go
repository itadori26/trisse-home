package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/meals/domain"
	"waylight/internal/modules/meals/usecase"
	"waylight/internal/shared/kernel"
)

type fakeUC struct {
	usecase.Meals
	from string
	days int
}

func (f *fakeUC) Range(_ context.Context, from string, days int) ([]usecase.EntryView, error) {
	f.from, f.days = from, days
	return nil, nil
}

func (f *fakeUC) Create(_ context.Context, in domain.Input, visibility string) (usecase.EntryView, error) {
	e, err := domain.New("1", in, kernel.Access{Owner: "anna", Visibility: kernel.Visibility(visibility)}, time.Time{})
	return usecase.EntryView{Entry: e, Mine: true}, err
}

func TestList(t *testing.T) {
	f := &fakeUC{}
	rec := httptest.NewRecorder()
	NewHandlers(f).List(rec, httptest.NewRequest(http.MethodGet, "/api/meals?from=2026-11-23", nil))
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" || f.from != "2026-11-23" || f.days != defaultDays {
		t.Fatalf("%d %s %+v", rec.Code, rec.Body, f)
	}
	rec = httptest.NewRecorder()
	NewHandlers(f).List(rec, httptest.NewRequest(http.MethodGet, "/api/meals?days=x", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("days=x: %d", rec.Code)
	}
}

func TestCreate(t *testing.T) {
	rec := httptest.NewRecorder()
	body := `{"date":"2026-11-23","meal":"dinner","dish":"Soup","ingredients":["leeks"],"visibility":"everyone"}`
	NewHandlers(&fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/meals", strings.NewReader(body)))
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusCreated || got["dish"] != "Soup" || got["visibility"] != "everyone" || got["mine"] != true {
		t.Fatalf("%d %v", rec.Code, got)
	}
	rec = httptest.NewRecorder()
	NewHandlers(&fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/meals", strings.NewReader(`{"date":"2026-11-23","dish":""}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty dish: %d", rec.Code)
	}
}
