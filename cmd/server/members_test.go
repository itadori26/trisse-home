package main

import (
	"net/http"
	"testing"
)

// TestMembersSwitch: members are on by default (#294); with the switch off
// the server refuses to create them, which the auth use cases still support.
func TestMembersSwitch(t *testing.T) {
	saved := membersEnabled
	membersEnabled = false
	t.Cleanup(func() { membersEnabled = saved })
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var list obj
	owner.json(http.MethodGet, "/api/auth/members", nil, http.StatusOK, &list)
	if list["available"] != false {
		t.Fatalf("members must be off with the switch off: %v", list)
	}
	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusConflict)

	// The interface learns who owns each module's data (ADR 0018).
	var info struct {
		DataKinds map[string]string `json:"dataKinds"`
	}
	owner.json(http.MethodGet, "/api/modules", nil, http.StatusOK, &info)
	if info.DataKinds["notes"] != "personal-content" || info.DataKinds["appearance"] != "personal-settings" || info.DataKinds["services"] != "instance" {
		t.Fatalf("data kinds: %v", info.DataKinds)
	}
}

// TestMemberAccess checks the guard for members (#294) with members switched
// on: they change personal content and settings, read instance data, and
// manage nothing of the instance.
func TestMemberAccess(t *testing.T) {
	saved := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = saved })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var member obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusCreated, &member)
	var group obj
	owner.json(http.MethodPost, "/api/auth/groups", obj{"name": "Office", "accounts": []string{member["id"].(string)}}, http.StatusCreated, &group)

	m := newClient(t, srv.URL)
	var ms obj
	m.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "member-password-1"}, http.StatusOK, &ms)
	m.csrf = ms["csrfToken"].(string)
	var status obj
	m.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["role"] != "member" || status["username"] != "ben" {
		t.Fatalf("status: %v", status)
	}

	// Personal content and settings: read and write.
	m.do(http.MethodGet, "/api/notes", nil, http.StatusOK)
	var memberNote, ownerNote obj
	m.json(http.MethodPost, "/api/notes", obj{"title": "Mine", "body": "member secret"}, http.StatusCreated, &memberNote)
	owner.json(http.MethodPost, "/api/notes", obj{"title": "Family", "body": "for everyone"}, http.StatusCreated, &ownerNote)

	// Notes are owned per record (ADR 0018, #294 step 3): the member's note
	// is private, the owner's new note is for everyone as before, and only
	// the owner of a note can change it.
	var ownerNotes, memberNotes []obj
	owner.json(http.MethodGet, "/api/notes", nil, http.StatusOK, &ownerNotes)
	for _, n := range ownerNotes {
		if n["id"] == memberNote["id"] {
			t.Fatal("the owner sees the member's private note")
		}
	}
	owner.do(http.MethodGet, "/api/notes/"+memberNote["id"].(string), nil, http.StatusNotFound)
	owner.do(http.MethodPut, "/api/notes/"+memberNote["id"].(string), obj{"title": "x", "body": "y"}, http.StatusNotFound)
	m.json(http.MethodGet, "/api/notes", nil, http.StatusOK, &memberNotes)
	if len(memberNotes) < 2 {
		t.Fatalf("member sees its own and the shared note: %v", memberNotes)
	}
	m.do(http.MethodGet, "/api/notes/"+ownerNote["id"].(string), nil, http.StatusOK)
	m.do(http.MethodPut, "/api/notes/"+ownerNote["id"].(string), obj{"title": "Hijacked", "body": "y"}, http.StatusForbidden)
	m.do(http.MethodDelete, "/api/notes/"+ownerNote["id"].(string), nil, http.StatusForbidden)
	// To-dos follow the same rules.
	var memberTodo, ownerTodo obj
	m.json(http.MethodPost, "/api/todos", obj{"title": "Member task"}, http.StatusCreated, &memberTodo)
	owner.json(http.MethodPost, "/api/todos", obj{"title": "Family task"}, http.StatusCreated, &ownerTodo)
	var ownerTodos []obj
	owner.json(http.MethodGet, "/api/todos", nil, http.StatusOK, &ownerTodos)
	for _, td := range ownerTodos {
		if td["id"] == memberTodo["id"] {
			t.Fatal("the owner sees the member's private to-do")
		}
	}
	m.do(http.MethodPut, "/api/todos/"+ownerTodo["id"].(string), obj{"title": "Hijacked"}, http.StatusForbidden)
	m.do(http.MethodDelete, "/api/todos/"+ownerTodo["id"].(string), nil, http.StatusForbidden)
	owner.do(http.MethodDelete, "/api/todos/"+memberTodo["id"].(string), nil, http.StatusNotFound)

	// The admin chooses private for their own new records (ADR 0018,
	// migration step 3); from the next request on, members no longer see
	// them, while older shared notes stay visible.
	var prefs obj
	owner.json(http.MethodGet, "/api/auth/preferences", nil, http.StatusOK, &prefs)
	if prefs["privateByDefault"] != false {
		t.Fatalf("preferences: %v", prefs)
	}
	owner.do(http.MethodPut, "/api/auth/preferences", obj{"privateByDefault": true}, http.StatusOK)
	m.do(http.MethodGet, "/api/auth/preferences", nil, http.StatusForbidden)
	var diary obj
	owner.json(http.MethodPost, "/api/notes", obj{"title": "Admin diary", "body": "private now"}, http.StatusCreated, &diary)
	m.do(http.MethodGet, "/api/notes/"+diary["id"].(string), nil, http.StatusNotFound)
	m.do(http.MethodGet, "/api/notes/"+ownerNote["id"].(string), nil, http.StatusOK)

	var hits []obj
	owner.json(http.MethodGet, "/api/notes/search?q=secret", nil, http.StatusOK, &hits)
	if len(hits) != 0 {
		t.Fatalf("the owner found the member's note: %v", hits)
	}
	m.do(http.MethodGet, "/api/appearance", nil, http.StatusOK)
	m.do(http.MethodGet, "/api/modules", nil, http.StatusOK)

	// Instance data: read only; instance settings and admin areas: closed.
	m.do(http.MethodGet, "/api/services", nil, http.StatusOK)
	m.do(http.MethodPost, "/api/services", obj{"name": "x", "url": "http://x"}, http.StatusForbidden)
	for _, path := range []string{"/api/auth/guests", "/api/auth/members", "/api/auth/groups", "/api/auth/events", "/api/auth/tokens",
		"/api/transfer/export", "/api/backups", "/api/backup/compact", "/api/notifications/settings", "/api/adguard/settings", "/api/docker/containers"} {
		m.do(http.MethodGet, path, nil, http.StatusForbidden)
	}
	m.do(http.MethodPost, "/api/auth/members", obj{"name": "eve", "password": "member-password-2"}, http.StatusForbidden)
	m.do(http.MethodPost, "/api/auth/guests", obj{"name": "eve", "password": "guest-password-2"}, http.StatusForbidden)

	// Self-service on the own account works.
	m.do(http.MethodGet, "/api/auth/sessions", nil, http.StatusOK)

	// Settings that act for the instance stay with the admin (#294 step 5).
	m.do(http.MethodGet, "/api/parcels/mail", nil, http.StatusForbidden)
	m.do(http.MethodPut, "/api/parcels/settings", obj{"dhlApiKey": "x", "hideDeliveredDays": 3}, http.StatusForbidden)
	m.do(http.MethodPut, "/api/birthdays/settings", obj{"remindEnabled": true, "remindAt": "08:00"}, http.StatusForbidden)
	m.do(http.MethodPut, "/api/bookmarks/settings", obj{"enrich": true, "checkLinks": true}, http.StatusForbidden)

	// Personal tabs (#294 step 2): the member gets its own dashboard and
	// cannot see or change the owner's tabs.
	var odash, mdash struct {
		Tabs []obj `json:"tabs"`
	}
	owner.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &odash)
	ownerTab := odash.Tabs[0]["id"].(string)
	m.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &mdash)
	for _, tab := range mdash.Tabs {
		if tab["id"] == ownerTab {
			t.Fatalf("member sees the owner's tab: %v", mdash.Tabs)
		}
	}
	var mine obj
	m.json(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Ben"}, http.StatusCreated, &mine)
	m.do(http.MethodPost, "/api/dashboard/tabs/"+mine["id"].(string)+"/widgets", obj{"kind": "clock", "x": 0, "y": 0, "w": 4, "h": 2}, http.StatusCreated)
	m.do(http.MethodPost, "/api/dashboard/tabs/"+ownerTab+"/widgets", obj{"kind": "clock", "x": 0, "y": 0, "w": 4, "h": 2}, http.StatusNotFound)
	m.do(http.MethodPut, "/api/dashboard/tabs/"+ownerTab, obj{"name": "Mine"}, http.StatusNotFound)
	m.do(http.MethodDelete, "/api/dashboard/tabs/"+ownerTab, nil, http.StatusNotFound)
	owner.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &odash)
	for _, tab := range odash.Tabs {
		if tab["id"] == mine["id"] {
			t.Fatal("the owner sees the member's personal tab")
		}
	}

	// The security log records the account and group changes.
	var events []obj
	owner.json(http.MethodGet, "/api/auth/events", nil, http.StatusOK, &events)
	seen := map[string]bool{}
	for _, e := range events {
		seen[e["type"].(string)] = true
	}
	if !seen["member_changed"] || !seen["group_changed"] {
		t.Fatalf("events: %v", seen)
	}

	// Disabling signs the member out at once.
	owner.do(http.MethodPut, "/api/auth/members/"+member["id"].(string), obj{"disabled": true}, http.StatusOK)
	m.do(http.MethodGet, "/api/notes", nil, http.StatusUnauthorized)
	owner.do(http.MethodDelete, "/api/auth/members/"+member["id"].(string), nil, http.StatusNoContent)
	owner.do(http.MethodDelete, "/api/auth/groups/"+group["id"].(string), nil, http.StatusNoContent)
}

// TestMembersOnByDefault: a fresh instance can create members (#294).
func TestMembersOnByDefault(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var list obj
	owner.json(http.MethodGet, "/api/auth/members", nil, http.StatusOK, &list)
	if list["available"] != true {
		t.Fatalf("members must be on: %v", list)
	}
	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusCreated)
}
