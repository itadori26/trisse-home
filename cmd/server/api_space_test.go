package main

import (
	"io"
	"net/http"
	"time"

	"trisse-home/internal/platform/config"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "space",
		Upstream: func(mux *http.ServeMux) {
			// Launch Library 2 and NASA APOD (#64).
			mux.HandleFunc("GET /ll2/", func(w http.ResponseWriter, _ *http.Request) {
				net := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
				_, _ = io.WriteString(w, `{"results":[{"id":"1","name":"Falcon 9 | Starlink 12-3","net":"`+net+`","status":{"abbrev":"Go"},`+
					`"rocket":{"configuration":{"name":"Falcon 9"}},"launch_service_provider":{"name":"SpaceX"},"pad":{"name":"SLC-40","location":{"name":"Cape Canaveral"}}}]}`)
			})
			mux.HandleFunc("GET /apod", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"date":"2026-11-13","title":"A Video","media_type":"video","url":"https://www.youtube.com/embed/x"}`)
			})
		},
		Config: func(cfg *config.Config, upstream string) {
			cfg.SpaceLaunchAPIURL, cfg.SpacePictureAPIURL = upstream+"/ll2/", upstream+"/apod"
		},
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			// Off by default: nothing is requested.
			c.do(http.MethodGet, "/api/space/launches", nil, http.StatusNotFound)
			c.json(http.MethodPut, "/api/space/settings", obj{"launches": true, "picture": true}, http.StatusOK, nil)
			var launches []any
			c.json(http.MethodGet, "/api/space/launches?limit=3", nil, http.StatusOK, &launches)
			if len(launches) != 1 || launches[0].(obj)["rocket"] != "Falcon 9" || launches[0].(obj)["mission"] != "Starlink 12-3" {
				t.Fatalf("launches: %v", launches)
			}
			var pic obj
			c.json(http.MethodGet, "/api/space/apod", nil, http.StatusOK, &pic)
			if pic["mediaType"] != "video" || pic["image"] != "" {
				t.Fatalf("apod: %v", pic)
			}
			c.do(http.MethodGet, "/api/space/apod/image", nil, http.StatusNotFound)
		},
	})
}
