package spa

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// gzExt marks the gzip-compressed variant of a file. The frontend build stores
// compressible files only as name.gz (web/vite.config.ts) so the binary
// carries about a third of their size; files without a .gz neighbour (images,
// fonts) are served as they are.
const gzExt = ".gz"

// Handler serves files from fsys. Hashed assets below /assets/ are cached
// forever; index.html is never cached so new releases are picked up at once.
// A file stored as name.gz is sent as is to clients that accept gzip and
// decompressed for the others.
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
		if !exists(fsys, name) || name == "index.html" || strings.HasSuffix(name, gzExt) {
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
		if serveCompressed(w, r, fsys, name) {
			return
		}
		files.ServeHTTP(w, r)
	})
}

// exists reports whether name or its gzip variant is in fsys.
func exists(fsys fs.FS, name string) bool {
	if _, err := fs.Stat(fsys, name+gzExt); err == nil {
		return true
	}
	_, err := fs.Stat(fsys, name)
	return !errors.Is(err, fs.ErrNotExist)
}

// serveCompressed serves name from name.gz and reports whether there was one.
func serveCompressed(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) bool {
	packed, err := fs.ReadFile(fsys, name+gzExt)
	if err != nil {
		return false
	}
	w.Header().Add("Vary", "Accept-Encoding")
	body := packed
	if acceptsGzip(r) {
		w.Header().Set("Content-Encoding", "gzip")
	} else if body, err = gunzip(packed); err != nil {
		http.Error(w, "frontend file is damaged", http.StatusInternalServerError)
		return true
	}
	// name (not name.gz) decides the content type; http.ServeContent also answers
	// HEAD, conditional and range requests.
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
	return true
}

// acceptsGzip reports whether the client lists gzip with a non-zero weight.
func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		coding, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(strings.TrimSpace(coding), "gzip") {
			continue
		}
		q := strings.ReplaceAll(strings.ToLower(params), " ", "")
		return q != "q=0" && q != "q=0.0" && q != "q=0.00" && q != "q=0.000"
	}
	return false
}

func gunzip(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	return io.ReadAll(zr)
}

// readFile reads name, or name.gz decompressed.
func readFile(fsys fs.FS, name string) ([]byte, error) {
	if packed, err := fs.ReadFile(fsys, name+gzExt); err == nil {
		return gunzip(packed)
	}
	return fs.ReadFile(fsys, name)
}

func serveIndex(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	body, err := readFile(fsys, "index.html")
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
