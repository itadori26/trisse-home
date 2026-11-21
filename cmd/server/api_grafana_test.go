package main

import (
	"io"
	"net/http"
	"strings"
)

// End-to-end check of the Grafana module (#574): connect with a token, list
// dashboards and panels, render a panel and read the alerts; the token is
// never returned.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "grafana",
		Upstream: func(mux *http.ServeMux) {
			auth := func(h http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "Bearer glsa_e2e" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					h(w, r)
				}
			}
			mux.HandleFunc("GET /grafana/api/search", auth(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `[{"uid":"node","title":"Node Exporter","url":"/d/node/node","folderTitle":"Servers"}]`)
			}))
			mux.HandleFunc("GET /grafana/api/dashboards/uid/node", auth(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"dashboard":{"panels":[{"id":2,"title":"CPU","type":"timeseries"}]}}`)
			}))
			mux.HandleFunc("GET /grafana/render/d-solo/node/_", auth(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = io.WriteString(w, "\x89PNG")
			}))
			mux.HandleFunc("GET /grafana/api/prometheus/grafana/api/v1/alerts", auth(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"data":{"alerts":[{"labels":{"alertname":"Disk full"},"annotations":{"summary":"95 %"},"state":"Alerting","activeAt":"2026-11-20T08:00:00Z"}]}}`)
			}))
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			c.do(http.MethodGet, "/api/grafana/alerts", nil, http.StatusNotFound)
			// A widget that chose a named connection reads through it (#610);
			// an unknown one falls back to the (still empty) module settings.
			var conn obj
			c.json(http.MethodPost, "/api/integrations/connections", obj{"service": "grafana", "name": "Upstream", "url": up.URL + "/grafana",
				"credential": obj{"token": "glsa_e2e"}}, http.StatusCreated, &conn)
			id := conn["id"].(string)
			c.do(http.MethodPost, "/api/integrations/connections/"+id+"/check", nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/grafana/dashboards/node/panels?connection="+id, nil, http.StatusOK)
			c.do(http.MethodGet, "/api/grafana/alerts?connection="+id, nil, http.StatusOK)
			c.do(http.MethodGet, "/api/grafana/alerts?connection=missing", nil, http.StatusNotFound)
			c.do(http.MethodDelete, "/api/integrations/connections/"+id, nil, http.StatusNoContent)
			c.do(http.MethodPut, "/api/grafana/settings", obj{"url": up.URL + "/grafana", "token": "wrong"}, http.StatusUnprocessableEntity)
			var settings obj
			c.json(http.MethodPut, "/api/grafana/settings", obj{"url": up.URL + "/grafana", "token": "glsa_e2e"}, http.StatusOK, &settings)
			if settings["hasToken"] != true || settings["token"] != nil {
				t.Fatalf("grafana settings: %v", settings)
			}
			var dashboards []obj
			c.json(http.MethodGet, "/api/grafana/dashboards?q=node", nil, http.StatusOK, &dashboards)
			if len(dashboards) != 1 || dashboards[0]["uid"] != "node" || !strings.HasSuffix(dashboards[0]["url"].(string), "/grafana/d/node/node") {
				t.Fatalf("grafana dashboards: %v", dashboards)
			}
			var panels []obj
			c.json(http.MethodGet, "/api/grafana/dashboards/node/panels", nil, http.StatusOK, &panels)
			if len(panels) != 1 || panels[0]["title"] != "CPU" {
				t.Fatalf("grafana panels: %v", panels)
			}
			if png := c.do(http.MethodGet, "/api/grafana/render/node/2?range=6h", nil, http.StatusOK); string(png) != "\x89PNG" {
				t.Fatalf("grafana render: %q", png)
			}
			c.do(http.MethodGet, "/api/grafana/render/node/2?range=now-6h", nil, http.StatusUnprocessableEntity)
			var alerts obj
			c.json(http.MethodGet, "/api/grafana/alerts", nil, http.StatusOK, &alerts)
			list := alerts["alerts"].([]any)
			if len(list) != 1 || list[0].(obj)["state"] != "firing" || !strings.HasSuffix(alerts["url"].(string), "/alerting/list") {
				t.Fatalf("grafana alerts: %v", alerts)
			}
			var answers struct {
				Answers []obj `json:"answers"`
			}
			c.json(http.MethodGet, "/api/search/answers?q=grafana+alarme&lang=de", nil, http.StatusOK, &answers)
			if len(answers.Answers) == 0 || answers.Answers[0]["title"] != "Grafana-Alarme: 1" {
				t.Fatalf("grafana answer: %v", answers)
			}
		},
	})
}
