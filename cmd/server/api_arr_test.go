package main

import (
	"io"
	"net/http"
	"strings"
	"time"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "arr",
		Upstream: func(mux *http.ServeMux) {
			// Sonarr (#86).
			mux.HandleFunc("GET /sonarr/api/v3/system/status", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"appName":"Sonarr"}`)
			})
			mux.HandleFunc("GET /sonarr/api/v3/calendar", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `[{"title":"Pilot","seasonNumber":1,"episodeNumber":1,"airDateUtc":"`+time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339)+`","series":{"title":"Show"}}]`)
			})
			mux.HandleFunc("GET /sonarr/api/v3/queue", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `{"records":[]}`) })
			mux.HandleFunc("GET /sonarr/api/v3/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `[]`) })
			mux.HandleFunc("GET /sonarr/api/v3/wanted/missing", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"totalRecords":2}`)
			})
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			_ = up
			c.do(http.MethodGet, "/api/arr/overview", nil, http.StatusNotFound)
			c.do(http.MethodPost, "/api/arr/instances", obj{"app": "radarr", "url": up.URL + "/sonarr", "apiKey": "k"}, http.StatusUnprocessableEntity)
			c.json(http.MethodPost, "/api/arr/instances", obj{"app": "sonarr", "url": up.URL + "/sonarr", "apiKey": "sonarr-key"}, http.StatusCreated, nil)
			if body := string(c.do(http.MethodGet, "/api/arr/instances", nil, http.StatusOK)); strings.Contains(body, "sonarr-key") {
				t.Fatal("arr API key leaked")
			}
			var arrOverview obj
			c.json(http.MethodGet, "/api/arr/overview", nil, http.StatusOK, &arrOverview)
			if up := arrOverview["upcoming"].([]any); len(up) != 1 || arrOverview["missing"].(obj)["Sonarr"] != float64(2) {
				t.Fatalf("arr: %v", arrOverview)
			}
		},
	})
}
