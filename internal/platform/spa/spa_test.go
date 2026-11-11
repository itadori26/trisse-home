package spa

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandler(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":           {Data: []byte("<html>app</html>")},
		"assets/app-1.js":      {Data: []byte("js")},
		"sw.js":                {Data: []byte("worker")},
		"manifest.webmanifest": {Data: []byte(`{"name":"x"}`)},
	}
	h := Handler(fsys)
	tests := []struct {
		path, wantBody, wantCache string
	}{
		{"/", "app", "no-cache"},
		{"/settings/deep/link", "app", "no-cache"},
		{"/assets/app-1.js", "js", "immutable"},
		{"/sw.js", "worker", "no-cache"},
		{"/manifest.webmanifest", "name", "no-cache"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if !strings.Contains(rec.Body.String(), tt.wantBody) || !strings.Contains(rec.Header().Get("Cache-Control"), tt.wantCache) {
			t.Errorf("%s: body=%q cache=%q", tt.path, rec.Body.String(), rec.Header().Get("Cache-Control"))
		}
	}
}

func TestManifestContentType(t *testing.T) {
	h := Handler(fstest.MapFS{"index.html": {Data: []byte("x")}, "manifest.webmanifest": {Data: []byte("{}")}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/manifest+json" {
		t.Fatalf("content type %q", ct)
	}
}
