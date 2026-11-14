package main

import "net/http"

func init() {
	registerModuleE2E(moduleE2E{
		Name: "summary",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var s obj
			c.json(http.MethodGet, "/api/summary/settings", nil, http.StatusOK, &s)
			if s["configured"] != false {
				t.Fatalf("summary: %v", s)
			}
			// Hidden until configured; only Ollama in the home network.
			c.do(http.MethodPost, "/api/summary", obj{"text": "Ein ausreichend langer Text für eine Zusammenfassung."}, http.StatusNotFound)
			c.do(http.MethodPost, "/api/summary/ask", obj{"question": "? Warum war Jellyfin offline?", "context": "", "language": "de"}, http.StatusNotFound)
			c.do(http.MethodPut, "/api/summary/settings", obj{"url": "https://api.example.com", "model": "m"}, http.StatusUnprocessableEntity)
			c.do(http.MethodPost, "/api/summary/models", obj{"url": "http://localhost:1"}, http.StatusUnprocessableEntity)
			c.do(http.MethodDelete, "/api/summary/settings", nil, http.StatusNoContent)
		},
	})
}
