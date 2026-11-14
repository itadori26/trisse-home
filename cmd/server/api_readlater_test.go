package main

import (
	"io"
	"net/http"
	"strings"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "readlater",
		Upstream: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /readlater/post", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = io.WriteString(w, `<html><head><title>Internal</title></head><body><article><p>secret</p></article></body></html>`)
			})
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			// The upstream runs on loopback: the guarded client must refuse it
			// (#99 SSRF guard); the link is kept with the reason.
			var a obj
			c.json(http.MethodPost, "/api/readlater", obj{"url": up.URL + "/readlater/post"}, http.StatusCreated, &a)
			if a["error"] == "" || a["content"] != nil || a["status"] != "unread" {
				t.Fatalf("loopback page must not be fetched: %v", a)
			}
			c.json(http.MethodPost, "/api/readlater", obj{"url": "http://169.254.169.254/latest/meta-data"}, http.StatusUnprocessableEntity, nil)
			id, _ := a["id"].(string)
			var list []obj
			c.json(http.MethodGet, "/api/readlater", nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["id"] != id {
				t.Fatalf("list: %v", list)
			}
			c.json(http.MethodPut, "/api/readlater/"+id+"/status", obj{"status": "archived"}, http.StatusNoContent, nil)
			c.json(http.MethodGet, "/api/readlater?status=archived", nil, http.StatusOK, &list)
			if len(list) != 1 {
				t.Fatalf("archived: %v", list)
			}
			got := string(c.do(http.MethodGet, "/api/readlater/"+id, nil, http.StatusOK))
			if strings.Contains(got, "secret") {
				t.Fatalf("content of a loopback page: %s", got)
			}
			c.do(http.MethodDelete, "/api/readlater/"+id, nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/readlater/"+id+"/images/0", nil, http.StatusNotFound)

			// Part 2: Pocket import (pending copies) and the cleanup setting.
			csv := "title,url,time_added,tags,status\nEins,https://a.example/1,1700000000,,unread\nZwei,https://a.example/2,1700000100,,archive\n"
			req, _ := http.NewRequest(http.MethodPost, c.base+"/api/readlater/import/pocket", strings.NewReader(csv))
			req.Header.Set("Content-Type", "text/csv")
			req.Header.Set("X-CSRF-Token", c.csrf)
			res, err := c.http.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"added":2`) {
				t.Fatalf("import: %d %s", res.StatusCode, body)
			}
			c.json(http.MethodGet, "/api/readlater", nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["pending"] != true || list[0]["title"] != "Eins" {
				t.Fatalf("imported: %v", list)
			}
			var s obj
			c.json(http.MethodGet, "/api/readlater/settings", nil, http.StatusOK, &s)
			if s["archiveDays"] != float64(0) {
				t.Fatalf("settings: %v", s)
			}
			c.do(http.MethodPut, "/api/readlater/settings", obj{"archiveDays": 400}, http.StatusUnprocessableEntity)
			c.json(http.MethodPut, "/api/readlater/settings", obj{"archiveDays": 30}, http.StatusOK, &s)
		},
	})
}
