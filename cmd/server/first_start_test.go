package main

import (
	"net/http"
	"testing"

	"waylight/internal/platform/config"
)

// TestFirstStartTemplateReplacesSample checks the first start with sample
// data (#670): the example tabs are marked, a template followed by removing
// the example tabs leaves only the template's zoned tab, and the module
// example data stays.
func TestFirstStartTemplateReplacesSample(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.SampleData = true })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	type dashboard struct {
		Tabs []struct {
			Name    string `json:"name"`
			Sample  bool   `json:"sample"`
			Zoned   bool   `json:"zoned"`
			Widgets []struct {
				Kind string `json:"kind"`
				Size string `json:"size"`
				Zone string `json:"zone"`
			} `json:"widgets"`
		} `json:"tabs"`
	}
	var dash dashboard
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 2 || !dash.Tabs[0].Sample || !dash.Tabs[1].Sample {
		t.Fatalf("example tabs not marked: %+v", dash.Tabs)
	}

	c.json(http.MethodPost, "/api/dashboard/templates/builtin/family/tabs",
		obj{"name": "Familie", "location": obj{"lat": 53.14, "lon": 8.21, "label": "Oldenburg"}}, http.StatusCreated, nil)
	c.do(http.MethodDelete, "/api/dashboard/sample", nil, http.StatusNoContent)
	dash = dashboard{}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 1 || dash.Tabs[0].Name != "Familie" || !dash.Tabs[0].Zoned || dash.Tabs[0].Sample {
		t.Fatalf("after the template: %+v", dash.Tabs)
	}
	zones := map[string]string{}
	for _, w := range dash.Tabs[0].Widgets {
		zones[w.Kind] = w.Zone + "/" + w.Size
	}
	if zones["weather"] != "today/small" || zones["services"] != "system/strip" || zones["shopping"] != "mine/medium" {
		t.Fatalf("zones and sizes: %v", zones)
	}

	var todos []obj
	c.json(http.MethodGet, "/api/todos", nil, http.StatusOK, &todos)
	if len(todos) == 0 {
		t.Fatal("module example data removed")
	}
}

// TestFirstStartKeepSample: keeping the example dashboard turns its tabs
// into ordinary tabs, so a later template choice cannot remove them (#670).
func TestFirstStartKeepSample(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.SampleData = true })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	c.do(http.MethodPost, "/api/dashboard/sample/keep", nil, http.StatusNoContent)
	c.do(http.MethodDelete, "/api/dashboard/sample", nil, http.StatusNoContent)
	var dash struct {
		Tabs []struct {
			Sample bool `json:"sample"`
		} `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 2 || dash.Tabs[0].Sample || dash.Tabs[1].Sample {
		t.Fatalf("kept tabs: %+v", dash.Tabs)
	}
}
