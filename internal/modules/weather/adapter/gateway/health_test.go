package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trisse-home/internal/modules/weather/domain"
)

// The fixtures are the real DWD files of 13 November 2026.
func TestHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/s31fg.json", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "testdata/s31fg.json") })
	mux.HandleFunc("/uvi.json", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "testdata/uvi.json") })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	h := NewHealth(NewDWD(srv.Client(), "", ""), srv.URL+"/s31fg.json", srv.URL+"/uvi.json")

	pollen, err := h.Pollen(context.Background())
	if err != nil || len(pollen) != 27 {
		t.Fatalf("%d %v", len(pollen), err)
	}
	var ni, mv *domain.PollenForecast
	for i := range pollen {
		switch pollen[i].Region.ID {
		case 32:
			ni = &pollen[i]
		case 20:
			mv = &pollen[i]
		}
	}
	if ni == nil || ni.Region.Name != "Niedersachsen und Bremen – Östl. Niedersachsen" || len(ni.Days) != 8 || ni.Days[0].Type != domain.Hazel {
		t.Fatalf("part region: %+v", ni)
	}
	// A region without part regions keeps its own id and name.
	if mv == nil || mv.Region.Name != "Mecklenburg-Vorpommern" {
		t.Fatalf("whole region: %+v", mv)
	}
	if got := ni.UpdatedAt.UTC().Format(time.RFC3339); got != "2026-11-13T10:00:00Z" {
		t.Fatalf("updated: %s", got)
	}

	uv, err := h.UV(context.Background())
	if err != nil || len(uv) != 38 {
		t.Fatalf("%d %v", len(uv), err)
	}
	if uv[0].City != "Bonn" || uv[0].Today != 1 || uv[1].City != "Freiburg" || uv[1].Today != 2 || uv[0].UpdatedAt.UTC().Hour() != 6 {
		t.Fatalf("%+v", uv[:2])
	}

	bad := httptest.NewServer(http.NotFoundHandler())
	defer bad.Close()
	if _, err := NewHealth(NewDWD(bad.Client(), "", ""), bad.URL, bad.URL).Pollen(context.Background()); err == nil {
		t.Fatal("404 must be an error")
	}
}
