package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/modules/calendar/usecase"
)

type fakeSubscriptions struct {
	usecase.Subscriptions
	added usecase.SubscriptionInput
}

func (f *fakeSubscriptions) ListSubscriptions(context.Context) ([]usecase.SubscriptionView, error) {
	return []usecase.SubscriptionView{{ID: "a", Name: "Uni", URLHint: "https://uni.example/…", Color: "#aa0000", RefreshMinutes: 60}}, nil
}

func (f *fakeSubscriptions) AddSubscription(_ context.Context, in usecase.SubscriptionInput) (usecase.SubscriptionView, error) {
	f.added = in
	return usecase.SubscriptionView{ID: "b", Name: in.Name}, nil
}

func (f *fakeSubscriptions) DeleteSubscription(_ context.Context, id string) error {
	if id != "a" {
		return domain.ErrSubscriptionNotFound
	}
	return nil
}

func TestSubscriptionHandlers(t *testing.T) {
	fake := &fakeSubscriptions{}
	h := NewSubscriptionHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/calendar/subscriptions", h.List)
	mux.HandleFunc("POST /api/calendar/subscriptions", h.Add)
	mux.HandleFunc("DELETE /api/calendar/subscriptions/{id}", h.Delete)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/calendar/subscriptions", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"urlHint":"https://uni.example/…"`) || strings.Contains(rec.Body.String(), `"url"`) {
		t.Fatalf("list must only expose the URL hint: %s", rec.Body)
	}

	rec = httptest.NewRecorder()
	body := `{"name":"Holidays","url":"webcal://h.example/de.ics","color":"#00aa00","refreshMinutes":1440}`
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/calendar/subscriptions", strings.NewReader(body)))
	if rec.Code != http.StatusCreated || fake.added.RefreshMinutes != 1440 || fake.added.URL != "webcal://h.example/de.ics" {
		t.Fatalf("add: %d %+v", rec.Code, fake.added)
	}

	for id, want := range map[string]int{"a": http.StatusNoContent, "zzz": http.StatusNotFound} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/calendar/subscriptions/"+id, nil))
		if rec.Code != want {
			t.Errorf("delete %s: got %d, want %d", id, rec.Code, want)
		}
	}
}
