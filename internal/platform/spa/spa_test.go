package spa

import (
	"bytes"
	"compress/gzip"
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

func gz(t *testing.T, s string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return buf.Bytes()
}

func TestCompressedFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html.gz":           {Data: gz(t, "<html>app</html>")},
		"assets/app-1.js.gz":      {Data: gz(t, "console.log(1)")},
		"sw.js.gz":                {Data: gz(t, "worker")},
		"manifest.webmanifest.gz": {Data: gz(t, `{"name":"x"}`)},
		"icons/icon.png":          {Data: []byte("png")},
	}
	h := Handler(fsys)
	get := func(path, acceptEncoding string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if acceptEncoding != "" {
			req.Header.Set("Accept-Encoding", acceptEncoding)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	t.Run("gzip clients get the stored bytes", func(t *testing.T) {
		rec := get("/assets/app-1.js", "br, gzip;q=0.8")
		if rec.Header().Get("Content-Encoding") != "gzip" || rec.Header().Get("Vary") != "Accept-Encoding" {
			t.Fatalf("headers: %v", rec.Header())
		}
		if !bytes.Equal(rec.Body.Bytes(), gz(t, "console.log(1)")) {
			t.Fatal("body is not the stored gzip data")
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
			t.Fatalf("content type %q", ct)
		}
		if !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
			t.Fatalf("cache %q", rec.Header().Get("Cache-Control"))
		}
	})
	for _, enc := range []string{"", "identity", "gzip;q=0", "deflate"} {
		t.Run("clients without gzip ("+enc+") get decompressed data", func(t *testing.T) {
			rec := get("/assets/app-1.js", enc)
			if rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "console.log(1)" {
				t.Fatalf("encoding=%q body=%q", rec.Header().Get("Content-Encoding"), rec.Body.String())
			}
		})
	}
	t.Run("index and deep links are served decompressed", func(t *testing.T) {
		for _, p := range []string{"/", "/index.html", "/settings/deep/link", "/index.html.gz"} {
			rec := get(p, "gzip")
			if rec.Body.String() != "<html>app</html>" || rec.Header().Get("Cache-Control") != "no-cache" {
				t.Errorf("%s: body=%q cache=%q", p, rec.Body.String(), rec.Header().Get("Cache-Control"))
			}
		}
	})
	t.Run("special headers survive", func(t *testing.T) {
		if rec := get("/sw.js", "gzip"); rec.Header().Get("Cache-Control") != "no-cache" || rec.Header().Get("Content-Encoding") != "gzip" {
			t.Fatalf("sw.js: %v", rec.Header())
		}
		if rec := get("/manifest.webmanifest", ""); rec.Header().Get("Content-Type") != "application/manifest+json" || rec.Body.String() != `{"name":"x"}` {
			t.Fatalf("manifest: %v %q", rec.Header(), rec.Body.String())
		}
	})
	t.Run("files without a gzip variant are served as they are", func(t *testing.T) {
		rec := get("/icons/icon.png", "gzip")
		if rec.Body.String() != "png" || rec.Header().Get("Content-Encoding") != "" {
			t.Fatalf("png: %v %q", rec.Header(), rec.Body.String())
		}
	})
	t.Run("HEAD works", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodHead, "/assets/app-1.js", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
			t.Fatalf("HEAD: %d, %d bytes", rec.Code, rec.Body.Len())
		}
	})
}
