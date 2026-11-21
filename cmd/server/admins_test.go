package main

import (
	"net/http"
	"testing"
)

// TestSeveralAdmins checks the rules of ADR 0022 (#434) through the API: an
// admin manages people and settings, keeps a private space of their own,
// cannot touch their own role, and leaves backups, the data transfer and API
// tokens to the owner.
func TestSeveralAdmins(t *testing.T) {
	saved := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = saved })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var anna, ben obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	annaID, benID := anna["id"].(string), ben["id"].(string)
	var promoted obj
	owner.json(http.MethodPut, "/api/auth/members/"+annaID, obj{"role": "admin"}, http.StatusOK, &promoted)
	if promoted["role"] != "admin" || promoted["roleSource"] != "" {
		t.Fatalf("promoted: %v", promoted)
	}

	a := newClient(t, srv.URL)
	var as obj
	a.json(http.MethodPost, "/api/auth/login", obj{"username": "anna", "password": "member-password-1"}, http.StatusOK, &as)
	a.csrf = as["csrfToken"].(string)
	var status obj
	a.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["role"] != "admin" {
		t.Fatalf("status: %v", status)
	}

	// Manages people and instance settings.
	var list obj
	a.json(http.MethodGet, "/api/auth/members", nil, http.StatusOK, &list)
	if n := len(list["members"].([]any)); n != 2 {
		t.Fatalf("members listed: %d", n)
	}
	a.do(http.MethodPost, "/api/auth/guests", obj{"name": "oma", "password": "guest-password-1"}, http.StatusCreated)
	a.do(http.MethodPost, "/api/auth/groups", obj{"name": "Office", "accounts": []string{benID}}, http.StatusCreated)
	a.do(http.MethodPost, "/api/services", obj{"name": "NAS", "url": "http://nas.local"}, http.StatusCreated)
	a.do(http.MethodGet, "/api/auth/events", nil, http.StatusOK)
	a.do(http.MethodPut, "/api/auth/members/"+benID, obj{"role": "admin"}, http.StatusOK)
	a.do(http.MethodPut, "/api/auth/members/"+benID, obj{"role": "member"}, http.StatusOK)

	// Not on themselves, and never on the owner.
	a.do(http.MethodPut, "/api/auth/members/"+annaID, obj{"role": "member"}, http.StatusForbidden)
	a.do(http.MethodPut, "/api/auth/members/"+annaID, obj{"disabled": true}, http.StatusForbidden)
	a.do(http.MethodDelete, "/api/auth/members/"+annaID, nil, http.StatusForbidden)
	a.do(http.MethodPut, "/api/auth/members/owner", obj{"disabled": true}, http.StatusNotFound)
	a.do(http.MethodDelete, "/api/auth/members/owner", nil, http.StatusNotFound)
	a.do(http.MethodPut, "/api/auth/members/"+annaID, obj{"role": "owner"}, http.StatusForbidden)
	a.do(http.MethodPut, "/api/auth/members/"+benID, obj{"role": "owner"}, http.StatusUnprocessableEntity)
	a.do(http.MethodPut, "/api/auth/members/"+benID, obj{"role": "guest"}, http.StatusUnprocessableEntity)

	// What contains or replaces everyone's data stays with the owner.
	for _, path := range []string{"/api/backups", "/api/backups/status", "/api/backup/compact", "/api/transfer/export", "/api/auth/tokens"} {
		a.do(http.MethodGet, path, nil, http.StatusForbidden)
	}
	a.do(http.MethodPost, "/api/auth/tokens", obj{"name": "x", "scopes": []string{"read"}}, http.StatusForbidden)

	// A private space of their own: the owner does not see anna's note.
	var note obj
	a.json(http.MethodPost, "/api/notes", obj{"title": "Mine", "body": "admin secret"}, http.StatusCreated, &note)
	owner.do(http.MethodGet, "/api/notes/"+note["id"].(string), nil, http.StatusNotFound)

	// Demoted by the owner: the session ends and admin pages are closed.
	owner.do(http.MethodPut, "/api/auth/members/"+annaID, obj{"role": "member"}, http.StatusOK)
	a.do(http.MethodGet, "/api/auth/members", nil, http.StatusUnauthorized)
	a2 := newClient(t, srv.URL)
	var as2 obj
	a2.json(http.MethodPost, "/api/auth/login", obj{"username": "anna", "password": "member-password-1"}, http.StatusOK, &as2)
	a2.do(http.MethodGet, "/api/auth/members", nil, http.StatusForbidden)
}
