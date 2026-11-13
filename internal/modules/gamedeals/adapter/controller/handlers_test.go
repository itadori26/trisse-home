package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/modules/gamedeals/usecase"
)

type fakeDeals struct {
	usecase.GameDeals
	saved domain.Settings
}

func (f *fakeDeals) SaveSettings(_ context.Context, s domain.Settings) (domain.Settings, error) {
	f.saved = s
	return s, nil
}

func (f *fakeDeals) Overview(context.Context) (domain.Overview, error) {
	return domain.Overview{Deals: []domain.Deal{{AppID: 620, Name: "Portal 2", DiscountPct: 80, EndsAt: time.Unix(0, 0)}, {AppID: 1, Name: "x", DiscountPct: 5}}}, nil
}

func TestHandlers(t *testing.T) {
	fake := &fakeDeals{}
	h := NewHandlers(fake)
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/gamedeals/settings", h.SaveSettings)
	mux.HandleFunc("GET /api/gamedeals/overview", h.Overview)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/gamedeals/settings", strings.NewReader(`{"steamId":"7656","country":"DE","minDiscount":30,"epic":true}`)))
	if rec.Code != http.StatusOK || fake.saved.MinDiscount != 30 || !fake.saved.Epic {
		t.Fatalf("save: %d %+v", rec.Code, fake.saved)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/gamedeals/overview", nil))
	body := rec.Body.String()
	if !strings.Contains(body, `"url":"https://store.steampowered.com/app/620"`) || !strings.Contains(body, `"endsAt":null`) || !strings.Contains(body, `"errors":[]`) {
		t.Fatalf("overview: %s", body)
	}
}
