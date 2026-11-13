package main

import (
	"io"
	"net/http"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "immich",
		Upstream: func(mux *http.ServeMux) {
			// Immich (#85).
			mux.HandleFunc("GET /immich/api/assets/statistics", func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("x-api-key") != "immich-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				_, _ = io.WriteString(w, `{"images":10,"videos":2}`)
			})
			mux.HandleFunc("GET /immich/api/server/version", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"major":1,"minor":120,"patch":0}`)
			})
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			_ = up
			c.do(http.MethodGet, "/api/immich/overview", nil, http.StatusNotFound)
			c.json(http.MethodPut, "/api/immich/settings", obj{"url": up.URL + "/immich", "apiKey": "immich-key"}, http.StatusOK, nil)
			var immich obj
			c.json(http.MethodGet, "/api/immich/overview", nil, http.StatusOK, &immich)
			if immich["photos"] != float64(10) || immich["version"] != "1.120.0" || len(immich["memories"].([]any)) != 0 {
				t.Fatalf("immich: %v", immich)
			}
		},
	})
}
