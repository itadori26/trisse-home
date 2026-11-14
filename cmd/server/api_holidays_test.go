package main

import (
	"io"
	"net/http"
	"time"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "holidays",
		Upstream: func(mux *http.ServeMux) {
			// An ICS file of school holidays (#55), next summer relative to today.
			mux.HandleFunc("GET /ferien/ni.ics", func(w http.ResponseWriter, _ *http.Request) {
				start := time.Now().AddDate(0, 0, 3)
				end := start.AddDate(0, 0, 14)
				_, _ = io.WriteString(w, "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:1\r\nSUMMARY:Sommerferien\r\n"+
					"DTSTART;VALUE=DATE:"+start.Format("20060102")+"\r\nDTEND;VALUE=DATE:"+end.Format("20060102")+"\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n")
			})
		},
		Check: func(e e2eEnv) {
			t, c, up := e.t, e.c, e.up
			var o obj
			c.json(http.MethodGet, "/api/holidays?state=BY&limit=30", nil, http.StatusOK, &o)
			if o["state"] != "BY" || o["school"] != false || len(o["entries"].([]any)) < 12 {
				t.Fatalf("public holidays: %v", o)
			}
			c.json(http.MethodPut, "/api/holidays/settings", obj{"schoolUrl": up.URL + "/missing.ics"}, http.StatusUnprocessableEntity, nil)
			c.json(http.MethodPut, "/api/holidays/settings", obj{"schoolUrl": up.URL + "/ferien/ni.ics"}, http.StatusOK, nil)
			c.json(http.MethodGet, "/api/holidays?state=NI&limit=3", nil, http.StatusOK, &o)
			found := false
			for _, e := range o["entries"].([]any) {
				found = found || e.(obj)["name"] == "Sommerferien"
			}
			if o["school"] != true || !found {
				t.Fatalf("school holidays: %v", o)
			}
			c.json(http.MethodPut, "/api/holidays/settings", obj{"schoolUrl": ""}, http.StatusOK, nil)
		},
	})
}
