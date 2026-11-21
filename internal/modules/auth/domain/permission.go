package domain

import (
	"slices"
	"strings"
)

// guestSelfService are the auth endpoints a guest may use for its own
// account (password, two-factor, sessions, logout, linked sign-ins) with any
// method.
var guestSelfService = []string{"/api/auth/logout", "/api/auth/password", "/api/auth/totp", "/api/auth/sessions", "/api/auth/app-sessions", "/api/auth/passkeys", "/api/auth/my-groups", "/api/auth/email", "/api/auth/identities", "/api/auth/microsoft"}

// guestReadable are read endpoints every guest needs to render the page.
// /api/search is the web search of the search field, which is part of every
// page since the search widget was removed (#221); it only redirects to the
// search engine, the search settings stay closed.
var guestReadable = []string{"/api/dashboard", "/api/modules", "/api/appearance", "/api/auth/status", "/api/search"}

// GuestReach is what a guest's tabs open up: the modules whose widgets are on
// them, minus the routes a module reserves for a widget kind that is not
// there (#168), e.g. the call list without a calls widget.
type GuestReach struct {
	Modules map[string]bool
	// Closed are path prefixes the guest may not read despite its modules.
	Closed []string
}

// GuestMayAccess decides whether a guest may call method on path. The
// modules of reach have their read endpoints allowed, except settings and the
// closed prefixes. Everything else is denied — the rule is an allow-list, so
// new endpoints are closed for guests by default.
func GuestMayAccess(method, path string, reach GuestReach) bool {
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
	for _, p := range reach.Closed {
		if path == p || strings.HasPrefix(path, p+"/") {
			return false
		}
	}
	parts := strings.SplitN(strings.TrimPrefix(path, "/api/"), "/", 2)
	if !strings.HasPrefix(path, "/api/") || parts[0] == "" {
		return false
	}
	switch parts[0] {
	case "auth", "dashboard", "transfer", "backups", "backup", "notifications", "backupmonitor":
		return false
	}
	return reach.Modules[parts[0]]
}

// Data kinds of modules as the composition root reports them (ADR 0018). The
// values match modules.DataKind; auth's domain does not import the module
// registry.
const (
	KindPersonalContent  = "personal-content"
	KindPersonalSettings = "personal-settings"
)

// adminOnly are the API prefixes a member may not use at all, not even to
// read: account management, the instance export and backups, notification
// channels (they hold webhook secrets) and the device of a Pi appliance.
var adminOnly = []string{"auth", "transfer", "backups", "backup", "backupmonitor", "notifications", "device", "docker", "system"}

// MemberMayAccess decides whether a member may call method on path (#294).
// kinds maps module names (the first path segment after /api/) to their data
// kind. A member
//
//   - uses the self-service endpoints of its own account like a guest,
//   - reads every module except the admin-only ones and settings pages of
//     instance modules (which may name hosts and keys),
//   - writes only to modules with personal content or personal settings.
//
// Everything else is denied, so a new module without a kind stays closed for
// writes by members.
func MemberMayAccess(method, path string, kinds map[string]string) bool {
	for _, p := range guestSelfService {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	if path == "/api/auth/status" && (method == "GET" || method == "HEAD") {
		return true
	}
	if !strings.HasPrefix(path, "/api/") {
		return false
	}
	module, _, _ := strings.Cut(strings.TrimPrefix(path, "/api/"), "/")
	if module == "" || slices.Contains(adminOnly, module) {
		return false
	}
	kind := kinds[module]
	personal := kind == KindPersonalContent || kind == KindPersonalSettings
	if method == "GET" || method == "HEAD" {
		return personal || !strings.Contains(path, "/settings")
	}
	return personal
}

// ModuleOfPath is the module an API path belongs to: "/api/notes/1" → "notes".
func ModuleOfPath(path string) string {
	module, _, _ := strings.Cut(strings.TrimPrefix(path, "/api/"), "/")
	return module
}

// SecondFactorSetupPath reports what an account may reach while an admin
// policy requires a second factor it has not set up yet (#304): its own
// account endpoints (status, two-factor and passkey setup, sign-out) and what
// the interface needs to show the setup page.
func SecondFactorSetupPath(method, path string) bool {
	if strings.HasPrefix(path, "/api/auth/") {
		return true
	}
	read := method == "GET" || method == "HEAD"
	return read && (path == "/api/modules" || path == "/api/appearance")
}

// ownerOnlyPrefixes are the routes only the owner may use, also against
// admins (ADR 0022): backups and the data transfer contain or replace the
// data of everyone, including private records and the owner's account.
// "/api/backup" is the compact state of the backup widget (#677).
var ownerOnlyPrefixes = []string{"/api/backups", "/api/backup", "/api/transfer"}

// OwnerOnlyPath reports whether path is one of the owner's alone.
func OwnerOnlyPath(path string) bool {
	for _, p := range ownerOnlyPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}
