package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/modules/bookmarks/usecase"
)

type fakeIcon struct{}

func (fakeIcon) Execute(_ context.Context, id string) (domain.Icon, error) {
	if id != "b1" {
		return domain.Icon{}, domain.ErrIconNotFound
	}
	return domain.Icon{ContentType: "image/png", Data: []byte{1, 2}}, nil
}

type fakeSettings struct{ saved usecase.SettingsView }

func (f *fakeSettings) Execute(_ context.Context, s usecase.SettingsView) (usecase.SettingsView, error) {
	f.saved = s
	return s, nil
}

type fakeCheck struct{}

func (fakeCheck) Execute(_ context.Context, id string) (usecase.BookmarkView, error) {
	return usecase.BookmarkView{ID: id, Link: domain.LinkStatus{State: domain.LinkBroken, Code: 404,
		CheckedAt: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}}, nil
}

func TestLinkAndIconHandlers(t *testing.T) {
	list, settings := &fakeList{}, &fakeSettings{}
	h := NewHandlers(UseCases{List: list, Icon: fakeIcon{}, SaveSettings: settings, CheckLink: fakeCheck{}})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/bookmarks", h.List)
	mux.HandleFunc("GET /api/bookmarks/{id}/icon", h.Icon)
	mux.HandleFunc("PUT /api/bookmarks/settings", h.SaveSettings)
	mux.HandleFunc("POST /api/bookmarks/{id}/check", h.CheckLink)
	do := func(method, path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
		return rec
	}

	if rec := do(http.MethodGet, "/api/bookmarks?status=broken", ""); rec.Code != http.StatusOK || !list.got.Broken {
		t.Fatalf("broken filter not passed: %+v", list.got)
	}
	rec := do(http.MethodGet, "/api/bookmarks/b1/icon", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" ||
		rec.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.Contains(rec.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Fatalf("icon %d %v", rec.Code, rec.Header())
	}
	if rec := do(http.MethodGet, "/api/bookmarks/b2/icon", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing icon %d", rec.Code)
	}
	if rec := do(http.MethodPut, "/api/bookmarks/settings", `{"enrich":true,"checkLinks":false}`); rec.Code != http.StatusOK || !settings.saved.Enrich {
		t.Fatalf("settings %d %+v", rec.Code, settings.saved)
	}
	rec = do(http.MethodPost, "/api/bookmarks/b1/check", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"link":{"state":"broken","code":404,"target":"","checkedAt":"2026-11-12T10:00:00Z"}`) {
		t.Fatalf("check %d %s", rec.Code, rec.Body)
	}
}
