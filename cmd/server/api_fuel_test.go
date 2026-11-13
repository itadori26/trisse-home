package main

import (
	"io"
	"net/http"
	"trisse-home/internal/platform/config"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "fuel",
		Upstream: func(mux *http.ServeMux) {
			// Tankerkönig (#53).
			mux.HandleFunc("GET /tk/list.php", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"ok":true,"stations":[{"id":"a","name":"JET","place":"Berlin","dist":1.2,"price":1.739,"isOpen":true}]}`)
			})
		},
		Config: func(cfg *config.Config, upstream string) {
			cfg.FuelAPIURL = upstream + "/tk"
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			_ = up
			c.do(http.MethodGet, "/api/fuel/overview", nil, http.StatusNotFound)
			c.json(http.MethodPut, "/api/fuel/settings", obj{"apiKey": "00000000-0000-0000-0000-000000000002", "lat": 52.5, "lon": 13.4,
				"place": "Berlin", "radiusKm": 5, "fuel": "e10"}, http.StatusOK, nil)
			var fuel obj
			c.json(http.MethodGet, "/api/fuel/overview", nil, http.StatusOK, &fuel)
			if cheapest := fuel["cheapest"].([]any); len(cheapest) != 1 || cheapest[0].(obj)["price"] != 1.739 {
				t.Fatalf("fuel: %v", fuel)
			}
		},
	})
}
