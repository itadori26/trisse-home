package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/platform/database/dbtest"
)

// fakeDAV is a minimal WebDAV collection at /dav/ with basic auth.
func fakeDAV(t *testing.T) (*httptest.Server, map[string][]byte) {
	t.Helper()
	var mu sync.Mutex
	files := map[string][]byte{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || u != "me" || p != "pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/dav/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/dav/")
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
			fmt.Fprint(w, `<?xml version="1.0"?><d:multistatus xmlns:d="DAV:"><d:response><d:href>/dav/</d:href><d:propstat><d:prop/></d:propstat></d:response>`)
			if r.Header.Get("Depth") == "1" {
				for n, b := range files {
					fmt.Fprintf(w, `<d:response><d:href>/dav/%s</d:href><d:propstat><d:prop><d:getcontentlength>%d</d:getcontentlength><d:getlastmodified>Wed, 11 Nov 2026 03:00:00 GMT</d:getlastmodified></d:prop></d:propstat></d:response>`, n, len(b))
				}
			}
			fmt.Fprint(w, `</d:multistatus>`)
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			files[name] = b
			w.WriteHeader(http.StatusCreated)
		case http.MethodGet:
			b, ok := files[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write(b)
		case http.MethodDelete:
			delete(files, name)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, files
}

func TestWebDAV(t *testing.T) {
	srv, files := fakeDAV(t)
	ctx := context.Background()
	dav := NewWebDAV(srv.Client())
	s := domain.RemoteSettings{URL: srv.URL + "/dav/", Username: "me", Password: "pw", Retention: 3}
	if err := dav.Test(ctx, s); err != nil || len(files) != 0 {
		t.Fatalf("test: %v (left %d files)", err, len(files))
	}
	bad := s
	bad.Password = "wrong"
	if err := dav.Test(ctx, bad); err == nil || !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("wrong password: %v", err)
	}
	missing := s
	missing.URL = srv.URL + "/nope/"
	if err := dav.Test(ctx, missing); err == nil {
		t.Fatal("missing folder must fail")
	}
	name := "trisse-home-20261111-030000.db"
	if err := dav.Put(ctx, s, name, strings.NewReader("sqlite"), 6); err != nil {
		t.Fatal(err)
	}
	files["notes.txt"] = []byte("ignored")
	list, err := dav.List(ctx, s)
	if err != nil || len(list) != 1 || list[0].Name != name || list[0].Size != 6 || list[0].CreatedAt.Hour() != 3 {
		t.Fatalf("list: %+v %v", list, err)
	}
	rc, err := dav.Get(ctx, s, name)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(b) != "sqlite" {
		t.Fatalf("get: %q", b)
	}
	if _, err := dav.Get(ctx, s, "../etc/passwd"); !errors.Is(err, domain.ErrBackupNotFound) {
		t.Fatalf("traversal: %v", err)
	}
	if err := dav.Delete(ctx, s, name); err != nil || len(files) != 1 {
		t.Fatalf("delete: %v %d", err, len(files))
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestRemoteSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "backup", Migrations())
	r := NewRemoteSettingsRepository(db, prefixCipher{})
	if _, err := r.Load(ctx); !errors.Is(err, domain.ErrRemoteNotConfigured) {
		t.Fatal(err)
	}
	_ = r.Save(ctx, domain.RemoteSettings{URL: "https://c/dav/", Username: "me", Password: "pw", Retention: 7})
	var stored string
	_ = db.QueryRow(`SELECT password_encrypted FROM backup_remote`).Scan(&stored)
	s, err := r.Load(ctx)
	if err != nil || s.Password != "pw" || s.Retention != 7 || stored != "enc:pw" {
		t.Fatalf("%+v %v %q", s, err, stored)
	}
	_ = r.Delete(ctx)
	if _, err := r.Load(ctx); !errors.Is(err, domain.ErrRemoteNotConfigured) {
		t.Fatal("delete")
	}
}
