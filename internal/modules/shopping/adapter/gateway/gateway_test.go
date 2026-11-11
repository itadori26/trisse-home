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

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/platform/database/dbtest"
)

// fakeTaskList is a minimal CalDAV collection with ETags.
type fakeTaskList struct {
	mu      sync.Mutex
	objects map[string]string
	version map[string]int
}

func (f *fakeTaskList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, p, _ := r.BasicAuth(); u != "me" || p != "pw" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/tasks/")
	etag := func(n string) string { return fmt.Sprintf(`"%d"`, f.version[n]) }
	switch r.Method {
	case "REPORT":
		w.WriteHeader(http.StatusMultiStatus)
		var b strings.Builder
		b.WriteString(`<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">`)
		for n, data := range f.objects {
			fmt.Fprintf(&b, `<d:response><d:href>/tasks/%s</d:href><d:propstat><d:prop><d:getetag>%s</d:getetag><c:calendar-data>%s</c:calendar-data></d:prop></d:propstat></d:response>`, n, etag(n), data)
		}
		b.WriteString(`</d:multistatus>`)
		_, _ = io.WriteString(w, b.String())
	case http.MethodGet:
		data, ok := f.objects[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", etag(name))
		_, _ = io.WriteString(w, data)
	case http.MethodPut:
		_, exists := f.objects[name]
		if r.Header.Get("If-None-Match") == "*" && exists {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		if m := r.Header.Get("If-Match"); m != "" && m != etag(name) {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		f.objects[name] = string(body)
		f.version[name]++
		w.WriteHeader(http.StatusCreated)
	case http.MethodDelete:
		delete(f.objects, name)
		w.WriteHeader(http.StatusNoContent)
	}
}

func TestCalDAVStore(t *testing.T) {
	ctx := context.Background()
	list := &fakeTaskList{objects: map[string]string{}, version: map[string]int{}}
	srv := httptest.NewServer(list)
	defer srv.Close()
	s := domain.Settings{ListURL: srv.URL + "/tasks/", Username: "me", Password: "pw"}
	store := NewCalDAVStore(srv.Client())
	it, err := store.Add(ctx, s, "u1@trisse-home", "Milk")
	if err != nil || it.ID != "u1-trisse-home.ics" {
		t.Fatalf("add: %+v %v", it, err)
	}
	if err := store.SetDone(ctx, s, it.ID, true); err != nil {
		t.Fatal(err)
	}
	items, err := store.List(ctx, s)
	if err != nil || len(items) != 1 || !items[0].Done || items[0].Summary != "Milk" {
		t.Fatalf("list: %+v %v", items, err)
	}
	if err := store.SetDone(ctx, s, "missing.ics", true); !errors.Is(err, domain.ErrItemNotFound) {
		t.Fatalf("missing item: %v", err)
	}
	if err := store.Delete(ctx, s, it.ID); err != nil {
		t.Fatal(err)
	}
	s.Password = "wrong"
	if _, err := store.List(ctx, s); err == nil {
		t.Fatal("wrong password accepted")
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepository(dbtest.Open(t, "shopping", Migrations()), prefixCipher{})
	_ = repo.Save(ctx, domain.Settings{ListURL: "https://c/tasks/", Username: "me", Password: "pw"})
	if s, err := repo.Load(ctx); err != nil || s.Password != "pw" {
		t.Fatalf("%+v %v", s, err)
	}
}
