package main

import "net/http"

func init() {
	registerModuleE2E(moduleE2E{
		Name: "study",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			c.json(http.MethodPut, "/api/study/settings", obj{"totalEcts": 210}, http.StatusOK, nil)
			var m obj
			c.json(http.MethodPost, "/api/study/modules", obj{"name": "Analysis I", "ects": 10, "semester": 1, "status": "passed", "grade": 1.0}, http.StatusCreated, &m)
			c.json(http.MethodPost, "/api/study/modules", obj{"name": "Lineare Algebra", "ects": 5, "semester": 1, "status": "passed", "grade": 2.5}, http.StatusCreated, nil)
			c.json(http.MethodPost, "/api/study/modules", obj{"name": "Programmieren", "ects": 8, "semester": 2, "status": "in_progress"}, http.StatusCreated, nil)
			c.json(http.MethodPost, "/api/study/modules", obj{"name": "x", "ects": 5, "status": "open", "grade": 2.0}, http.StatusUnprocessableEntity, nil)
			var p obj
			c.json(http.MethodGet, "/api/study/progress", nil, http.StatusOK, &p)
			if p["passedEcts"] != 15.0 || p["average"] != 1.5 || p["totalEcts"] != 210.0 || len(p["inProgress"].([]any)) != 1 {
				t.Fatalf("progress: %v", p)
			}
			c.do(http.MethodDelete, "/api/study/modules/"+m["id"].(string), nil, http.StatusNoContent)
		},
	})
}
