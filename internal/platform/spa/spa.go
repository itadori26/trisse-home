package spa

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Handler serves files from fsys. Hashed assets below /assets/ are cached
// forever; index.html is never cached so new releases are picked up at once.
func Handler(fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(fsys, name); errors.Is(err, fs.ErrNotExist) || name == "index.html" {
			serveIndex(w, r, fsys)
			return
		}
		switch {
		case strings.HasPrefix(name, "assets/"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case name == "sw.js":
			// The service worker must always be revalidated, or updates never arrive.
			w.Header().Set("Cache-Control", "no-cache")
		case strings.HasSuffix(name, ".webmanifest"):
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Type", "application/manifest+json")
		}
		files.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	body, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.Error(w, "frontend not built", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodGet {
		_, _ = w.Write(body)
	}
}
