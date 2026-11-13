package main

import (
	"io"
	"net/http"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "apiwidget",
		Upstream: func(mux *http.ServeMux) {
			// A home-lab JSON API behind an API key (#48).
			mux.HandleFunc("GET /jsonapi/status", func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Api-Key") != "k3y" {
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				_, _ = io.WriteString(w, `{"disks":[{"temp":48},{"temp":31}],"state":"healthy"}`)
			})
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			endpoint := obj{"name": "NAS", "url": up.URL + "/jsonapi/status", "headerName": "X-Api-Key", "headerValue": "k3y",
				"intervalSeconds": 60, "extractions": []obj{
					{"label": "Disk 1", "path": "$.disks[0].temp", "unit": "°C", "warn": 45, "crit": 55},
					{"label": "State", "path": "state"},
				}}
			var preview obj
			c.json(http.MethodPost, "/api/apiwidget/preview", obj{"endpoint": endpoint}, http.StatusOK, &preview)
			if vs := preview["values"].([]any); len(vs) != 2 || vs[0].(obj)["status"] != "warning" || vs[1].(obj)["text"] != "healthy" {
				t.Fatalf("preview: %v", preview)
			}
			var created obj
			c.json(http.MethodPost, "/api/apiwidget/endpoints", endpoint, http.StatusCreated, &created)
			id, _ := created["id"].(string)
			if created["hasHeader"] != true || created["headerValue"] != nil {
				t.Fatalf("created: %v", created)
			}
			var values obj
			c.json(http.MethodGet, "/api/apiwidget/endpoints/"+id+"/values", nil, http.StatusOK, &values)
			if values["error"] != "" || len(values["values"].([]any)) != 2 {
				t.Fatalf("values: %v", values)
			}
			// Saving without a new header value keeps the stored key.
			endpoint["headerValue"], endpoint["keepHeader"] = "", true
			c.json(http.MethodPut, "/api/apiwidget/endpoints/"+id, endpoint, http.StatusOK, nil)
			c.json(http.MethodPost, "/api/apiwidget/preview", obj{"id": id, "endpoint": endpoint}, http.StatusOK, &preview)
			if preview["error"] != "" {
				t.Fatalf("kept key: %v", preview)
			}
			c.do(http.MethodDelete, "/api/apiwidget/endpoints/"+id, nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/apiwidget/endpoints/"+id+"/values", nil, http.StatusNotFound)
		},
	})
}
