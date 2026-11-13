package main

import "net/http"

func init() {
	registerModuleE2E(moduleE2E{
		Name: "drop",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var item obj
			c.json(http.MethodPost, "/api/drop/text", obj{"text": "https://example.org/x", "expiresInHours": 1}, http.StatusCreated, &item)
			if item["kind"] != "link" {
				t.Fatalf("link: %v", item)
			}
			c.json(http.MethodPost, "/api/drop/text", obj{"text": "x", "expiresInHours": 5}, http.StatusUnprocessableEntity, nil)
			var list []obj
			c.json(http.MethodGet, "/api/drop", nil, http.StatusOK, &list)
			if len(list) != 1 {
				t.Fatalf("list: %v", list)
			}
			id, _ := item["id"].(string)
			c.do(http.MethodGet, "/api/drop/"+id+"/file", nil, http.StatusNotFound)
			c.do(http.MethodDelete, "/api/drop/"+id, nil, http.StatusNoContent)
		},
	})
}
