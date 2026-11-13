package main

import (
	"net/http"
	"testing"
	"time"

	"trisse-home/internal/platform/config"
)

// TestBriefingReadsOtherModules checks the composition-root adapters of
// the briefing (ADR 0016) end to end: data entered in the to-do module
// shows up in GET /api/briefing, and disabled source modules simply leave
// their sections out.
func TestBriefingReadsOtherModules(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) {
		// No outbound weather/calendar/waste requests in this test.
		c.DisabledModules = []string{"weather", "calendar", "waste"}
	})
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	today := time.Now().Format(time.DateOnly)
	yesterday := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
	nextWeek := time.Now().AddDate(0, 0, 7).Format(time.DateOnly)
	for _, td := range []obj{{"title": "Pay rent", "due": yesterday}, {"title": "Water plants", "due": today}, {"title": "Later", "due": nextWeek}} {
		c.do(http.MethodPost, "/api/todos", td, http.StatusCreated)
	}

	var b struct {
		Sections []struct {
			Kind  string `json:"kind"`
			Items []struct {
				Text     string `json:"text"`
				Detail   string `json:"detail"`
				Severity string `json:"severity"`
			} `json:"items"`
		} `json:"sections"`
		Unavailable []string `json:"unavailable"`
	}
	c.json(http.MethodGet, "/api/briefing", nil, http.StatusOK, &b)
	var todos []string
	for _, s := range b.Sections {
		if s.Kind == "events" || s.Kind == "waste" || s.Kind == "warnings" {
			t.Fatalf("disabled module produced section %s", s.Kind)
		}
		if s.Kind == "todos" {
			for _, it := range s.Items {
				todos = append(todos, it.Text)
				if it.Text == "Pay rent" && (it.Detail != "overdue" || it.Severity != "warning") {
					t.Fatalf("overdue to-do %+v", it)
				}
			}
		}
	}
	if len(todos) != 2 || todos[0] != "Pay rent" || todos[1] != "Water plants" {
		t.Fatalf("todos %v (sections %+v)", todos, b.Sections)
	}

	var settings struct {
		Available []string `json:"available"`
		Layout    []obj    `json:"layout"`
		PushOn    bool     `json:"pushOn"`
	}
	c.json(http.MethodGet, "/api/briefing/settings", nil, http.StatusOK, &settings)
	if settings.PushOn || len(settings.Layout) != 6 {
		t.Fatalf("defaults %+v", settings)
	}
	for _, k := range settings.Available {
		if k == "warnings" || k == "events" || k == "waste" {
			t.Fatalf("disabled source offered: %v", settings.Available)
		}
	}
	c.do(http.MethodPut, "/api/briefing/settings", obj{"layout": []obj{{"kind": "todos", "enabled": false}}, "pushOn": true,
		"pushAt": "25:00", "pushDays": []int{}, "skipEmpty": true, "lat": 52.5, "lon": 13.4}, http.StatusUnprocessableEntity)
	c.do(http.MethodPut, "/api/briefing/settings", obj{"layout": []obj{{"kind": "todos", "enabled": false}}, "pushOn": true,
		"pushAt": "07:00", "pushDays": []int{1, 2, 3, 4, 5}, "skipEmpty": true, "lat": 52.5, "lon": 13.4}, http.StatusOK)
	c.json(http.MethodGet, "/api/briefing", nil, http.StatusOK, &b)
	for _, s := range b.Sections {
		if s.Kind == "todos" {
			t.Fatal("a disabled section must be hidden")
		}
	}
	// Notifications are enabled, so a test push is accepted.
	c.do(http.MethodPost, "/api/briefing/send", nil, http.StatusNoContent)
}

// TestBriefingIgnoresUnconfiguredSources checks that an enabled source
// module that has not been set up yet (no waste calendar, no CalDAV
// account) is not reported as unavailable.
func TestBriefingIgnoresUnconfiguredSources(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) {
		c.DisabledModules = []string{"weather"} // the only source that calls out on its own
	})
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var b struct {
		Unavailable []string `json:"unavailable"`
	}
	c.json(http.MethodGet, "/api/briefing", nil, http.StatusOK, &b)
	if len(b.Unavailable) != 0 {
		t.Fatalf("unconfigured sources reported as unavailable: %v", b.Unavailable)
	}
}
