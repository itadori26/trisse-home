package main

import (
	"net/http"
	"testing"
)

// TestGuestAccess checks the server-side guest restrictions per endpoint
// group: guests only read what is on their shared tabs.
func TestGuestAccess(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var dash struct {
		Tabs []obj `json:"tabs"`
	}
	owner.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	home := dash.Tabs[0]["id"].(string)
	var second obj
	owner.json(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Private"}, http.StatusCreated, &second)

	var guest obj
	owner.json(http.MethodPost, "/api/auth/guests", obj{"name": "anna", "password": "guest-password-1", "tabs": []string{home}}, http.StatusCreated, &guest)
	owner.do(http.MethodPost, "/api/auth/guests", obj{"name": "anna", "password": "guest-password-1", "tabs": []string{}}, http.StatusConflict)

	g := newClient(t, srv.URL)
	g.do(http.MethodPost, "/api/auth/login", obj{"password": "guest-password-1"}, http.StatusUnauthorized) // the owner has another password
	var gs obj
	g.json(http.MethodPost, "/api/auth/login", obj{"username": "anna", "password": "guest-password-1"}, http.StatusOK, &gs)
	g.csrf = gs["csrfToken"].(string)
	var status obj
	g.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["role"] != "guest" || status["username"] != "anna" {
		t.Fatalf("status: %v", status)
	}

	// Dashboard: only the shared tab.
	var gdash struct {
		Tabs []obj `json:"tabs"`
	}
	g.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &gdash)
	if len(gdash.Tabs) != 1 || gdash.Tabs[0]["id"] != home {
		t.Fatalf("guest tabs: %v", gdash.Tabs)
	}

	// Reads: allowed for modules on the shared tab (the default tab has a
	// search widget), denied for everything else.
	g.do(http.MethodGet, "/api/modules", nil, http.StatusOK)
	g.do(http.MethodGet, "/api/appearance", nil, http.StatusOK)
	for _, path := range []string{"/api/notes", "/api/todos", "/api/system", "/api/docker/containers", "/api/bookmarks",
		"/api/notifications/settings", "/api/adguard/settings", "/api/auth/events", "/api/auth/guests", "/api/transfer/export",
		"/api/backups", "/api/dashboard/templates"} {
		g.do(http.MethodGet, path, nil, http.StatusForbidden)
	}

	// Writes: always denied, including widget actions and dashboard edits.
	g.do(http.MethodPost, "/api/notes", obj{"title": "x"}, http.StatusForbidden)
	g.do(http.MethodPost, "/api/todos", obj{"title": "x"}, http.StatusForbidden)
	g.do(http.MethodPost, "/api/dashboard/tabs", obj{"name": "x"}, http.StatusForbidden)
	g.do(http.MethodPut, "/api/appearance", obj{}, http.StatusForbidden)
	g.do(http.MethodPost, "/api/auth/guests", obj{"name": "eve", "password": "guest-password-2"}, http.StatusForbidden)
	g.do(http.MethodPost, "/api/transfer/import", obj{}, http.StatusForbidden)

	// Self-service on the own account works.
	g.do(http.MethodGet, "/api/auth/sessions", nil, http.StatusOK)
	g.do(http.MethodGet, "/api/auth/totp", nil, http.StatusOK)

	// Sharing the second tab takes effect immediately; disabling ends the session.
	owner.do(http.MethodPut, "/api/auth/guests/"+guest["id"].(string), obj{"tabs": []string{home, second["id"].(string)}, "disabled": false}, http.StatusOK)
	g.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &gdash)
	if len(gdash.Tabs) != 2 {
		t.Fatalf("second tab not shared: %v", gdash.Tabs)
	}
	owner.do(http.MethodPut, "/api/auth/guests/"+guest["id"].(string), obj{"tabs": []string{home}, "disabled": true}, http.StatusOK)
	g.do(http.MethodGet, "/api/dashboard", nil, http.StatusUnauthorized)

	// The owner's own session is unaffected and sees everything.
	owner.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 2 {
		t.Fatalf("owner tabs: %v", dash.Tabs)
	}
}
