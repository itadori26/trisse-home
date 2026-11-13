package domain

import "testing"

func TestGuestMayAccess(t *testing.T) {
	reach := GuestReach{Modules: map[string]bool{"notes": true, "weather": true, "fritzbox": true},
		Closed: []string{"/api/fritzbox/calls"}}
	cases := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/api/dashboard", true},
		{"GET", "/api/notes", true},
		{"GET", "/api/notes/abc", true},
		{"GET", "/api/weather/forecast", true},
		{"PUT", "/api/auth/password", true},
		{"POST", "/api/auth/totp/setup", true},
		{"DELETE", "/api/auth/sessions/x", true},
		{"POST", "/api/auth/logout", true},
		{"POST", "/api/notes", false},
		{"DELETE", "/api/notes/abc", false},
		{"GET", "/api/todos", false},
		{"GET", "/api/weather/settings", false},
		{"GET", "/api/auth/events", false},
		{"GET", "/api/auth/guests", false},
		{"GET", "/api/dashboard/templates", false},
		{"PUT", "/api/dashboard/tabs/1", false},
		{"GET", "/api/transfer/export", false},
		{"GET", "/api/backups", false},
		{"GET", "/api/", false},
		{"GET", "/healthz", false},
		// A module's personal route stays closed without its widget (#168).
		{"GET", "/api/fritzbox/status", true},
		{"GET", "/api/fritzbox/calls", false},
		{"GET", "/api/fritzbox/calls/x", false},
		{"GET", "/api/fritzbox/callsign", true},
		// The web search of the search field is part of every page (#221).
		{"GET", "/api/search", true},
		{"GET", "/api/search/settings", false},
		{"PUT", "/api/search/bar", false},
		{"POST", "/api/search/shortcuts", false},
	}
	for _, c := range cases {
		if got := GuestMayAccess(c.method, c.path, reach); got != c.want {
			t.Errorf("%s %s: %v, want %v", c.method, c.path, got, c.want)
		}
	}
}
