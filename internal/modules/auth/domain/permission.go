package domain

import "strings"

// guestSelfService are the auth endpoints a guest may use for its own
// account (password, two-factor, sessions, logout) with any method.
var guestSelfService = []string{"/api/auth/logout", "/api/auth/password", "/api/auth/totp", "/api/auth/sessions", "/api/auth/passkeys"}

// guestReadable are read endpoints every guest needs to render the page.
// /api/search is the web search of the search field, which is part of every
// page since the search widget was removed (#221); it only redirects to the
// search engine, the search settings stay closed.
var guestReadable = []string{"/api/dashboard", "/api/modules", "/api/appearance", "/api/auth/status", "/api/search"}

// GuestMayAccess decides whether a guest may call method on path. modules
// are the modules whose widgets are on the guest's tabs: their read
// endpoints are allowed, except settings. Everything else is denied — the
// rule is an allow-list, so new endpoints are closed for guests by default.
func GuestMayAccess(method, path string, modules map[string]bool) bool {
	for _, p := range guestSelfService {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	if method != "GET" && method != "HEAD" {
		return false
	}
	for _, p := range guestReadable {
		if path == p {
			return true
		}
	}
	if strings.Contains(path, "/settings") {
		return false
	}
	parts := strings.SplitN(strings.TrimPrefix(path, "/api/"), "/", 2)
	if !strings.HasPrefix(path, "/api/") || parts[0] == "" {
		return false
	}
	switch parts[0] {
	case "auth", "dashboard", "transfer", "backups", "notifications", "backupmonitor":
		return false
	}
	return modules[parts[0]]
}
