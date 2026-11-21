package main

import (
	"net/http"
	"time"
)

// End-to-end check of the chore plan module against the logged-in test
// server (#683).
func init() {
	registerModuleE2E(moduleE2E{
		Name: "chores",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			today := time.Now().Format(time.DateOnly)
			var created obj
			c.json(http.MethodPost, "/api/chores", obj{"title": "Dishes", "everyDays": 1, "due": today, "assignees": []string{"Anna", "Ben"}}, http.StatusCreated, &created)
			if created["assignee"] != "Anna" || created["dueNow"] != true || created["visibility"] != "everyone" {
				t.Fatalf("chores: %v", created)
			}
			var compact obj
			c.json(http.MethodGet, "/api/chores/compact", nil, http.StatusOK, &compact)
			action, _ := compact["action"].(map[string]any)
			if action["id"] != "chores.done" {
				t.Fatalf("chores compact: %v", compact)
			}
			id, _ := created["id"].(string)
			var done obj
			c.json(http.MethodPost, "/api/chores/"+id+"/done", nil, http.StatusOK, &done)
			if done["assignee"] != "Ben" || done["lastDoneBy"] != "Anna" || done["dueNow"] != false {
				t.Fatalf("chores done: %v", done)
			}
			c.json(http.MethodPut, "/api/chores/"+id, obj{"title": "Dishes", "everyDays": 0, "due": today}, http.StatusUnprocessableEntity, nil)
			c.do(http.MethodDelete, "/api/chores/"+id, nil, http.StatusNoContent)
		},
	})
}
