package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/adapter/presenter"
	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/modules/auth/usecase"
)

type fakeLogin struct{ err error }

func (f fakeLogin) Execute(context.Context, usecase.LoginRequest) (usecase.SessionResponse, error) {
	return usecase.SessionResponse{SessionToken: "tok", CSRFToken: "csrf", ExpiresAt: time.Now().Add(time.Hour)}, f.err
}

type fakeAuth struct{ err error }

func (f fakeAuth) Execute(_ context.Context, req usecase.AuthenticateRequest) (usecase.AuthenticateResponse, error) {
	if f.err != nil || req.SessionToken != "tok" {
		return usecase.AuthenticateResponse{}, domain.ErrSessionInvalid
	}
	return usecase.AuthenticateResponse{CSRFToken: "csrf"}, nil
}

func TestLoginSetsHardenedCookie(t *testing.T) {
	h := NewHandlers(UseCases{Login: fakeLogin{}}, CookieConfig{Secure: false})
	rec := httptest.NewRecorder()
	h.Login(rec, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password":"x"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	cookie := rec.Result().Cookies()[0]
	if cookie.Value != "tok" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie not hardened: %+v", cookie)
	}
	var dto presenter.SessionDTO
	_ = json.NewDecoder(rec.Body).Decode(&dto)
	if dto.CSRFToken != "csrf" || strings.Contains(rec.Body.String(), "tok\"") {
		t.Fatalf("unexpected body %s", rec.Body)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h := NewHandlers(UseCases{Login: fakeLogin{err: domain.ErrInvalidCredentials}}, CookieConfig{})
	rec := httptest.NewRecorder()
	h.Login(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"password":"x"}`)))
	if rec.Code != http.StatusUnauthorized || len(rec.Result().Cookies()) != 0 {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestSecureCookieUsesHostPrefix(t *testing.T) {
	h := NewHandlers(UseCases{Login: fakeLogin{}}, CookieConfig{Secure: true})
	rec := httptest.NewRecorder()
	h.Login(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"password":"x"}`)))
	c := rec.Result().Cookies()[0]
	if c.Name != hostCookieName || !c.Secure {
		t.Fatalf("unexpected cookie %+v", c)
	}
}

func TestGuard(t *testing.T) {
	h := NewHandlers(UseCases{Authenticate: fakeAuth{}}, CookieConfig{})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	guarded := h.Guard(next)
	tests := []struct {
		name, method, cookie, csrf string
		want                       int
	}{
		{"no session", http.MethodGet, "", "", http.StatusUnauthorized},
		{"valid get", http.MethodGet, "tok", "", http.StatusNoContent},
		{"post without csrf", http.MethodPost, "tok", "", http.StatusForbidden},
		{"post wrong csrf", http.MethodPost, "tok", "nope", http.StatusForbidden},
		{"post with csrf", http.MethodPost, "tok", "csrf", http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/api/x", nil)
			if tt.cookie != "" {
				r.AddCookie(&http.Cookie{Name: plainCookieName, Value: tt.cookie})
			}
			if tt.csrf != "" {
				r.Header.Set(CSRFHeader, tt.csrf)
			}
			rec := httptest.NewRecorder()
			guarded.ServeHTTP(rec, r)
			if rec.Code != tt.want {
				t.Fatalf("status %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
