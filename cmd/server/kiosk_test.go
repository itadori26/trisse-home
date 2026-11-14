package main

import (
	"net/http"
	"testing"
)

// TestKioskAccess checks kiosk displays (#50): the link signs a device in
// read-only for its tabs, without any account self-service.
func TestKioskAccess(t *testing.T) {
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
	owner.do(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Private"}, http.StatusCreated)

	owner.do(http.MethodPost, "/api/auth/kiosks", obj{"name": "flur", "tabs": []string{home}, "rotateSeconds": 5}, http.StatusUnprocessableEntity)
	var created struct {
		Kiosk obj    `json:"kiosk"`
		Token string `json:"token"`
	}
	owner.json(http.MethodPost, "/api/auth/kiosks", obj{"name": "flur", "tabs": []string{home}, "rotateSeconds": 30, "nightFrom": "22:00", "nightTo": "06:00"}, http.StatusCreated, &created)
	if created.Token == "" || created.Kiosk["rotateSeconds"] != float64(30) {
		t.Fatalf("created: %+v", created)
	}
	var list []obj
	owner.json(http.MethodGet, "/api/auth/kiosks", nil, http.StatusOK, &list)
	var guests []obj
	owner.json(http.MethodGet, "/api/auth/guests", nil, http.StatusOK, &guests)
	if len(list) != 1 || len(guests) != 0 {
		t.Fatalf("kiosks %v, guests %v", list, guests)
	}

	k := newClient(t, srv.URL)
	k.do(http.MethodPost, "/api/auth/login", obj{"username": "flur", "password": "!kiosk"}, http.StatusUnauthorized)
	k.do(http.MethodPost, "/api/auth/kiosk/login", obj{"token": "wrong"}, http.StatusUnauthorized)
	var ks obj
	k.json(http.MethodPost, "/api/auth/kiosk/login", obj{"token": created.Token}, http.StatusOK, &ks)
	k.csrf = ks["csrfToken"].(string)
	var status obj
	k.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["role"] != "kiosk" {
		t.Fatalf("status: %v", status)
	}
	var kdash struct {
		Tabs []obj `json:"tabs"`
	}
	k.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &kdash)
	if len(kdash.Tabs) != 1 || kdash.Tabs[0]["id"] != home {
		t.Fatalf("kiosk tabs: %v", kdash.Tabs)
	}
	var opts obj
	k.json(http.MethodGet, "/api/auth/kiosk", nil, http.StatusOK, &opts)
	if opts["nightFrom"] != "22:00" || opts["rotateSeconds"] != float64(30) {
		t.Fatalf("options: %v", opts)
	}
	// No self-service, no writes, no management.
	for _, path := range []string{"/api/auth/sessions", "/api/auth/totp", "/api/auth/passkeys", "/api/auth/kiosks", "/api/auth/guests", "/api/notes"} {
		k.do(http.MethodGet, path, nil, http.StatusForbidden)
	}
	k.do(http.MethodPut, "/api/auth/password", obj{"currentPassword": "x", "newPassword": "y"}, http.StatusForbidden)
	k.do(http.MethodPut, "/api/appearance", obj{}, http.StatusForbidden)
	k.do(http.MethodPost, "/api/dashboard/tabs", obj{"name": "x"}, http.StatusForbidden)
	owner.do(http.MethodGet, "/api/auth/kiosk", nil, http.StatusNotFound) // options are for displays

	// A new link signs the device out; the old link no longer works.
	var renewed obj
	owner.json(http.MethodPost, "/api/auth/kiosks/"+created.Kiosk["id"].(string)+"/link", nil, http.StatusOK, &renewed)
	k.do(http.MethodGet, "/api/dashboard", nil, http.StatusUnauthorized)
	k.do(http.MethodPost, "/api/auth/kiosk/login", obj{"token": created.Token}, http.StatusUnauthorized)
	owner.do(http.MethodDelete, "/api/auth/kiosks/"+created.Kiosk["id"].(string), nil, http.StatusNoContent)
	k.do(http.MethodPost, "/api/auth/kiosk/login", obj{"token": renewed["token"]}, http.StatusUnauthorized)
}
