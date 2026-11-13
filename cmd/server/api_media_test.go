package main

import (
	"io"
	"net/http"
	"strings"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "media",
		Upstream: func(mux *http.ServeMux) {
			// Jellyfin (#61).
			mux.HandleFunc("GET /jf/System/Info", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"ServerName":"Home"}`)
			})
			mux.HandleFunc("GET /jf/Items", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"Items":[{"Id":"abc","Name":"Film","Type":"Movie"}]}`)
			})
			mux.HandleFunc("GET /jf/Sessions", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, `[]`) })
			mux.HandleFunc("GET /jf/Items/abc/Images/Primary", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = io.WriteString(w, "png")
			})
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			_ = up
			c.do(http.MethodGet, "/api/media/overview", nil, http.StatusNotFound)
			c.json(http.MethodPut, "/api/media/settings", obj{"provider": "jellyfin", "url": up.URL + "/jf", "token": "jf-key"}, http.StatusOK, nil)
			if body := string(c.do(http.MethodGet, "/api/media/settings", nil, http.StatusOK)); strings.Contains(body, "jf-key") {
				t.Fatal("media token leaked")
			}
			var media obj
			c.json(http.MethodGet, "/api/media/overview", nil, http.StatusOK, &media)
			if media["serverName"] != "Home" || len(media["recent"].([]any)) != 1 {
				t.Fatalf("media: %v", media)
			}
			if body := string(c.do(http.MethodGet, "/api/media/posters/abc", nil, http.StatusOK)); body != "png" {
				t.Fatalf("poster: %q", body)
			}
			c.do(http.MethodGet, "/api/media/posters/other", nil, http.StatusNotFound)
		},
	})
}
