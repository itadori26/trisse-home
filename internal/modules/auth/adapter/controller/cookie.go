package controller

import (
	"net/http"
	"time"
)

const (
	plainCookieName = "trisse_session"
	// hostCookieName uses the __Host- prefix, which browsers only accept for
	// Secure, host-only cookies with Path=/. It hardens the cookie once HTTPS
	// is in place.
	hostCookieName = "__Host-trisse_session"
)

// CookieConfig controls how the session cookie is issued.
type CookieConfig struct {
	Secure bool
}

func (c CookieConfig) name() string {
	if c.Secure {
		return hostCookieName
	}
	return plainCookieName
}

func (c CookieConfig) read(r *http.Request) string {
	ck, err := r.Cookie(c.name())
	if err != nil {
		return ""
	}
	return ck.Value
}

func (c CookieConfig) set(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // Secure is configurable because the app may run over plain HTTP in the LAN
		Name: c.name(), Value: token, Path: "/", Expires: expires,
		HttpOnly: true, Secure: c.Secure, SameSite: http.SameSiteStrictMode,
	})
}

func (c CookieConfig) clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{ //nolint:gosec // Secure is configurable because the app may run over plain HTTP in the LAN
		Name: c.name(), Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: c.Secure, SameSite: http.SameSiteStrictMode,
	})
}
