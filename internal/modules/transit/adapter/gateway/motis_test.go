package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMOTIS(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/geocode", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "STOP" || r.Header.Get("User-Agent") == "" {
			t.Errorf("query %v", r.URL.Query())
		}
		_, _ = io.WriteString(w, `[{"type":"STOP","id":"de-DELFI_de:03403:16750","name":"Oldenburg Hbf","lat":53.14,"lon":8.22}]`)
	})
	mux.HandleFunc("GET /api/v1/stoptimes", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"stopTimes":[
			{"place":{"departure":"2026-11-11T20:17:00Z","scheduledDeparture":"2026-11-11T20:09:00Z","track":"A"},"mode":"BUS","realTime":true,"headsign":"Bümmerstede","routeShortName":"304","cancelled":false},
			{"place":{"arrival":"2026-11-11T20:20:00Z"},"mode":"BUS","headsign":"End","routeShortName":"9"}]}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	m := NewMOTIS(srv.Client(), srv.URL+"/")
	stops, err := m.SearchStops(context.Background(), "Oldenburg")
	if err != nil || len(stops) != 1 || stops[0].Name != "Oldenburg Hbf" {
		t.Fatalf("%+v %v", stops, err)
	}
	ds, err := m.Departures(context.Background(), stops[0].ID, 20)
	if err != nil || len(ds) != 1 || ds[0].DelayMinutes() != 8 || ds[0].Mode != "bus" || ds[0].Track != "A" {
		t.Fatalf("%+v %v", ds, err)
	}
}
