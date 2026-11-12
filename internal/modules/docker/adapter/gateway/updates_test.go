package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

// fakeRegistry speaks the distribution API with a bearer token challenge.
func fakeRegistry(t *testing.T) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token":
			user, pass, ok := r.BasicAuth()
			if r.URL.Query().Get("scope") != "repository:me/app:pull" || r.URL.Query().Get("service") != "test" || (ok && (user != "me" || pass != "pw")) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			_, _ = w.Write([]byte(`{"token":"tok123"}`))
		case r.Header.Get("Authorization") != "Bearer tok123":
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+srv.URL+`/token",service="test",scope="repository:me/app:pull"`)
			w.WriteHeader(http.StatusUnauthorized)
		case r.URL.Path == "/v2/me/app/manifests/1.0.0" && r.Method == http.MethodHead:
			if !strings.Contains(r.Header.Get("Accept"), "application/vnd.oci.image.index.v1+json") {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			w.Header().Set("Docker-Content-Digest", "sha256:remote")
		case r.URL.Path == "/v2/me/app/tags/list":
			_, _ = w.Write([]byte(`{"name":"me/app","tags":["1.0.0","1.1.0","latest"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRegistryClient(t *testing.T) {
	srv := fakeRegistry(t)
	host := strings.TrimPrefix(srv.URL, "http://")
	reg := &Registry{client: srv.Client(), scheme: "http"}
	ref, _ := domain.ParseImageRef(host + "/me/app:1.0.0")
	ctx := context.Background()
	for _, cred := range []*usecase.RegistryCredential{nil, {Host: host, Username: "me", Password: "pw"}} {
		d, err := reg.Digest(ctx, ref, cred)
		if err != nil || d != "sha256:remote" {
			t.Fatalf("digest (%v): %q %v", cred != nil, d, err)
		}
	}
	tags, err := reg.Tags(ctx, ref, nil)
	if err != nil || len(tags) != 3 {
		t.Fatalf("tags: %v %v", tags, err)
	}
	if _, err := reg.Digest(ctx, ref, &usecase.RegistryCredential{Username: "me", Password: "wrong"}); err == nil {
		t.Fatal("wrong credentials must fail")
	}
	missing, _ := domain.ParseImageRef(host + "/me/app:9.9.9")
	if _, err := reg.Digest(ctx, missing, nil); err == nil {
		t.Fatal("missing tag must fail")
	}
}

func TestParseChallenge(t *testing.T) {
	scheme, p := parseChallenge(`Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:library/nginx:pull"`)
	if scheme != "Bearer" || p["realm"] != "https://auth.docker.io/token" || p["service"] != "registry.docker.io" || p["scope"] != "repository:library/nginx:pull" {
		t.Fatalf("%s %+v", scheme, p)
	}
}

func TestRunningImages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte(`[{"Names":["/web"],"Image":"nginx:1.27","ImageID":"sha256:n"},{"Names":["/web2"],"Image":"nginx:1.27","ImageID":"sha256:n"},{"Names":["/db"],"Image":"postgres:16","ImageID":"sha256:p"}]`))
		case "/images/sha256:n/json":
			_, _ = w.Write([]byte(`{"RepoDigests":["nginx@sha256:aaa"],"Config":{"Labels":{"org.opencontainers.image.source":"https://github.com/nginx/docker-nginx"}}}`))
		case "/images/sha256:p/json":
			_, _ = w.Write([]byte(`{"RepoDigests":[],"Config":{"Labels":null}}`))
		default:
			w.WriteHeader(http.StatusForbidden) // the proxy without IMAGES=1
		}
	}))
	defer srv.Close()
	images, err := NewAPI(srv.Client(), srv.URL).RunningImages(context.Background())
	if err != nil || len(images) != 2 || images[0].Ref != "nginx:1.27" || len(images[0].Containers) != 2 || images[0].RepoDigests[0] != "nginx@sha256:aaa" ||
		images[0].Labels[domain.SourceLabel] == "" {
		t.Fatalf("%+v %v", images, err)
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestUpdateStore(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "docker", Migrations())
	s := NewUpdateStore(db, prefixCipher{})
	if set, err := s.Load(ctx); err != nil || set.Enabled {
		t.Fatalf("defaults: %+v %v", set, err)
	}
	want := usecase.UpdateSettings{Enabled: true, SemverTags: true, Credentials: []usecase.RegistryCredential{{Host: "ghcr.io", Username: "me", Password: "pw"}}}
	if err := s.Save(ctx, want); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRow(`SELECT password_encrypted FROM docker_registry_credentials`).Scan(&stored)
	got, _ := s.Load(ctx)
	if !got.Enabled || !got.SemverTags || len(got.Credentials) != 1 || got.Credentials[0].Password != "pw" || stored != "enc:pw" {
		t.Fatalf("%+v %q", got, stored)
	}
	at := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	for _, ref := range []string{"a:1", "b:1"} {
		_ = s.Put(ctx, usecase.ImageUpdate{Ref: ref, Status: domain.UpdateAvailable, CheckedAt: at, NewerTag: "2"})
	}
	_ = s.Put(ctx, usecase.ImageUpdate{Ref: "a:1", Status: domain.UpToDate, CheckedAt: at})
	_ = s.Retain(ctx, []string{"a:1"})
	list, err := s.List(ctx)
	if err != nil || len(list) != 1 || list[0].Status != domain.UpToDate || !list[0].CheckedAt.Equal(at) {
		t.Fatalf("%+v %v", list, err)
	}
	_ = s.Retain(ctx, nil)
	if list, _ := s.List(ctx); len(list) != 0 {
		t.Fatal("retain nothing")
	}
}
