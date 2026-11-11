package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/weather/domain"
)

func TestDWD(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ows", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Query().Get("CQL_FILTER"), "POINT(53.1400 8.2100)") {
			t.Errorf("filter %q", r.URL.Query().Get("CQL_FILTER"))
		}
		_, _ = io.WriteString(w, `{"features":[{"properties":{"IDENTIFIER":"x","EVENT":"STURMBÖEN","HEADLINE":"Amtliche WARNUNG vor STURMBÖEN",
			"SEVERITY":"Moderate","ONSET":"2026-11-12T09:00:00Z","EXPIRES":"2026-11-13T17:00:00Z"}}]}`)
	})
	mux.HandleFunc("GET /wms", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("bbox") == "0.0000,0.0000,0.0000,0.0000" {
			_, _ = io.WriteString(w, `<?xml version="1.0"?><ServiceExceptionReport/>`)
			return
		}
		_, _ = w.Write(append(pngSignature, 1, 2, 3))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	d := NewDWD(srv.Client(), srv.URL+"/ows", srv.URL+"/wms")
	loc, _ := domain.NewLocation(53.14, 8.21)
	ws, err := d.Warnings(context.Background(), loc)
	if err != nil || len(ws) != 1 || ws[0].Severity != domain.SeverityModerate || ws[0].Expires.Day() != 13 {
		t.Fatalf("%+v %v", ws, err)
	}
	if png, err := d.Radar(context.Background(), loc.RadarBBox(), 640, 520); err != nil || len(png) != 11 {
		t.Fatalf("radar: %v", err)
	}
	if _, err := d.Radar(context.Background(), domain.BBox{}, 1, 1); err == nil {
		t.Fatal("XML error page accepted as image")
	}
}
