package main

import (
	"net/http"
	"time"
)

// End-to-end check of the meal plan module against the logged-in test
// server (#682).
func init() {
	registerModuleE2E(moduleE2E{
		Name: "meals",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			today := time.Now().Format(time.DateOnly)
			var created obj
			c.json(http.MethodPost, "/api/meals", obj{"date": today, "meal": "dinner", "dish": "Lasagne", "ingredients": []string{"pasta"}}, http.StatusCreated, &created)
			if created["visibility"] != "everyone" || created["mine"] != true {
				t.Fatalf("meals: %v", created)
			}
			var list []obj
			c.json(http.MethodGet, "/api/meals?from="+today+"&days=7", nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["dish"] != "Lasagne" {
				t.Fatalf("meals: %v", list)
			}
			var compact obj
			c.json(http.MethodGet, "/api/meals/compact", nil, http.StatusOK, &compact)
			if title, _ := compact["title"].(map[string]any); title["text"] != "Today: Lasagne" {
				t.Fatalf("meals compact: %v", compact)
			}
			id, _ := created["id"].(string)
			c.json(http.MethodPut, "/api/meals/"+id, obj{"date": today, "dish": ""}, http.StatusUnprocessableEntity, nil)
			c.do(http.MethodDelete, "/api/meals/"+id, nil, http.StatusNoContent)
		},
	})
}
