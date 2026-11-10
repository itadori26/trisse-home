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
		"index.html":      {Data: []byte("<html>app</html>")},
		"assets/app-1.js": {Data: []byte("js")},
	}
	h := Handler(fsys)
	tests := []struct {
		path, wantBody, wantCache string
	}{
		{"/", "app", "no-cache"},
		{"/settings/deep/link", "app", "no-cache"},
		{"/assets/app-1.js", "js", "immutable"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if !strings.Contains(rec.Body.String(), tt.wantBody) || !strings.Contains(rec.Header().Get("Cache-Control"), tt.wantCache) {
			t.Errorf("%s: body=%q cache=%q", tt.path, rec.Body.String(), rec.Header().Get("Cache-Control"))
		}
	}
}
