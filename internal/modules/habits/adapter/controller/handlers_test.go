package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/habits/domain"
	"trisse-home/internal/modules/habits/usecase"
)

type fakeUC struct{ usecase.Habits }

var today = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

func view(name string) usecase.HabitView {
	h, _ := domain.New("1", name, "", domain.Schedule{}, 0, today)
	return usecase.HabitView{Habit: h, Stats: domain.ComputeStats(h, map[domain.Day]bool{today: true}, today)}
}

func (fakeUC) Create(_ context.Context, in usecase.Input) (usecase.HabitView, error) {
	if _, err := domain.NewSchedule(in.Days); err != nil {
		return usecase.HabitView{}, err
	}
	if _, err := domain.New("1", in.Name, in.Emoji, domain.Schedule{}, 0, today); err != nil {
		return usecase.HabitView{}, err
	}
	return view(in.Name), nil
}

func (fakeUC) Toggle(_ context.Context, id, date string) (usecase.HabitView, error) {
	if id != "1" {
		return usecase.HabitView{}, domain.ErrNotFound
	}
	if date != "2026-11-12" {
		return usecase.HabitView{}, domain.ErrDayNotEditable
	}
	return view("Read"), nil
}

func TestCreate(t *testing.T) {
	for body, want := range map[string]int{
		`{"name":"Read","days":[1,3]}`: http.StatusCreated,
		`{"name":""}`:                  http.StatusUnprocessableEntity,
		`{"name":"Read","days":[8]}`:   http.StatusUnprocessableEntity,
	} {
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/habits", strings.NewReader(body)))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", body, rec.Code, want)
		}
	}
}

func TestToggle(t *testing.T) {
	for _, c := range []struct {
		id, body string
		want     int
	}{
		{"1", `{"date":"2026-11-12"}`, http.StatusOK},
		{"1", `{"date":"2026-11-01"}`, http.StatusUnprocessableEntity},
		{"2", `{"date":"2026-11-12"}`, http.StatusNotFound},
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/habits/"+c.id+"/toggle", strings.NewReader(c.body))
		req.SetPathValue("id", c.id)
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{}).Toggle(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s %s: got %d, want %d", c.id, c.body, rec.Code, c.want)
		}
		if c.want == http.StatusOK && !strings.Contains(rec.Body.String(), `"doneToday":true`) {
			t.Errorf("body %s", rec.Body)
		}
	}
}
