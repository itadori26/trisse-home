package main

import (
	"net/http"
)

// Optimistic concurrency and idempotent creates of to-dos and notes for the
// offline outbox (#80).
func init() {
	registerModuleE2E(moduleE2E{
		Name: "offline-sync",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var todo obj
			c.json(http.MethodPost, "/api/todos", obj{"title": "Offline", "id": "outbox-todo-0001"}, http.StatusCreated, &todo)
			if todo["id"] != "outbox-todo-0001" || todo["version"] != 1.0 {
				t.Fatalf("todo: %v", todo)
			}
			// Replaying the create returns the same to-do.
			var again obj
			c.json(http.MethodPost, "/api/todos", obj{"title": "Offline", "id": "outbox-todo-0001"}, http.StatusCreated, &again)
			if again["id"] != todo["id"] || again["version"] != 1.0 {
				t.Fatalf("replay: %v", again)
			}
			c.json(http.MethodPut, "/api/todos/outbox-todo-0001", obj{"title": "Online", "version": 1}, http.StatusOK, &todo)
			if todo["version"] != 2.0 {
				t.Fatalf("version: %v", todo)
			}
			c.json(http.MethodPut, "/api/todos/outbox-todo-0001", obj{"title": "Stale", "version": 1, "done": true}, http.StatusConflict, nil)
			c.do(http.MethodDelete, "/api/todos/outbox-todo-0001", nil, http.StatusNoContent)

			var note obj
			c.json(http.MethodPost, "/api/notes", obj{"body": "Offline", "id": "outbox-note-0001"}, http.StatusCreated, &note)
			c.json(http.MethodPut, "/api/notes/outbox-note-0001", obj{"body": "Edited", "version": 1}, http.StatusOK, &note)
			c.json(http.MethodPut, "/api/notes/outbox-note-0001", obj{"body": "Stale", "version": 1}, http.StatusConflict, nil)
			c.do(http.MethodDelete, "/api/notes/outbox-note-0001", nil, http.StatusNoContent)
		},
	})
}
