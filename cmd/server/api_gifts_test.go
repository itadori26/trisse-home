package main

import "net/http"

// End-to-end check of the gift idea module against the logged-in test
// server (#684); the gateway test proves that ideas stay private.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "gifts",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var created obj
			c.json(http.MethodPost, "/api/gifts", obj{"person": "Grandma", "title": "Audiobook"}, http.StatusCreated, &created)
			var list []obj
			c.json(http.MethodGet, "/api/gifts", nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["title"] != "Audiobook" {
				t.Fatalf("gifts: %v", list)
			}
			var compact obj
			c.json(http.MethodGet, "/api/gifts/compact", nil, http.StatusOK, &compact)
			if title, _ := compact["title"].(map[string]any); title["text"] != "1 gift idea" {
				t.Fatalf("gifts compact: %v", compact)
			}
			id, _ := created["id"].(string)
			var bought obj
			c.json(http.MethodPut, "/api/gifts/"+id, obj{"person": "Grandma", "title": "Audiobook", "bought": true}, http.StatusOK, &bought)
			if bought["bought"] != true {
				t.Fatalf("gifts bought: %v", bought)
			}
			c.do(http.MethodDelete, "/api/gifts/"+id, nil, http.StatusNoContent)
		},
	})
}
