package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeImages struct{ list []LocalImage }

func (f *fakeImages) RunningImages(context.Context) ([]LocalImage, error) { return f.list, nil }

type fakeRegistry struct {
	digests map[string]string
	tags    []string
	calls   int
	creds   []*RegistryCredential
}

func (f *fakeRegistry) Digest(_ context.Context, ref domain.ImageRef, cred *RegistryCredential) (string, error) {
	f.calls++
	f.creds = append(f.creds, cred)
	d, ok := f.digests[ref.String()]
	if !ok {
		return "", errors.New("404")
	}
	return d, nil
}

func (f *fakeRegistry) Tags(context.Context, domain.ImageRef, *RegistryCredential) ([]string, error) {
	return f.tags, nil
}

type memCache struct{ m map[string]ImageUpdate }

func (c *memCache) List(context.Context) ([]ImageUpdate, error) {
	out := []ImageUpdate{}
	for _, u := range c.m {
		out = append(out, u)
	}
	return out, nil
}
func (c *memCache) Put(_ context.Context, u ImageUpdate) error { c.m[u.Ref] = u; return nil }
func (c *memCache) Retain(_ context.Context, refs []string) error {
	for k := range c.m {
		if !slices.Contains(refs, k) {
			delete(c.m, k)
		}
	}
	return nil
}

type memSettings struct{ s UpdateSettings }

func (m *memSettings) Load(context.Context) (UpdateSettings, error) { return m.s, nil }
func (m *memSettings) Save(_ context.Context, s UpdateSettings) error {
	m.s = s
	return nil
}

type obsRec struct{ got []kernel.Observation }

func (o *obsRec) Observe(_ context.Context, ob kernel.Observation) { o.got = append(o.got, ob) }

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

func TestImageUpdates(t *testing.T) {
	ctx := context.Background()
	images := &fakeImages{list: []LocalImage{
		{Ref: "nginx:1.27.1", Containers: []string{"web"}, RepoDigests: []string{"nginx@sha256:old"},
			Labels: map[string]string{domain.SourceLabel: "https://github.com/nginx/nginx"}},
		{Ref: "ghcr.io/me/app:latest", Containers: []string{"app"}, RepoDigests: []string{"ghcr.io/me/app@sha256:same"}},
		{Ref: "redis@sha256:pinned", Containers: []string{"cache"}, RepoDigests: []string{"redis@sha256:pinned"}},
		{Ref: "my-build:dev", Containers: []string{"dev"}},
	}}
	reg := &fakeRegistry{digests: map[string]string{"nginx:1.27.1": "sha256:new", "ghcr.io/me/app:latest": "sha256:same"}, tags: []string{"1.27.1", "1.27.3", "1.28.0"}}
	cache := &memCache{m: map[string]ImageUpdate{"gone:1": {Ref: "gone:1"}}}
	settings := &memSettings{}
	obs := &obsRec{}
	clk := &movingClock{t: time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)}
	uc := NewImageUpdates(UpdateDeps{Images: images, Registry: reg, Cache: cache, Settings: settings, Sink: obs, Clock: clk})

	// Off by default: no registry request at all.
	if err := uc.Check(ctx); err != nil || reg.calls != 0 {
		t.Fatalf("disabled check: %v %d", err, reg.calls)
	}
	if _, err := uc.List(ctx); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("disabled list: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, UpdateSettingsInput{Enabled: true, SemverTags: true,
		Credentials: []RegistryCredential{{Host: "GHCR.io", Username: "me", Password: "token"}}}); err != nil {
		t.Fatal(err)
	}
	if err := uc.Check(ctx); err != nil {
		t.Fatal(err)
	}
	if reg.calls != 2 || reg.creds[0] == nil && reg.creds[1] == nil {
		t.Fatalf("registry calls: %d %+v", reg.calls, reg.creds)
	}
	list, _ := uc.List(ctx)
	if len(list) != 4 || list[0].Ref != "nginx:1.27.1" || list[0].Status != domain.UpdateAvailable || list[0].NewerTag != "1.28.0" ||
		list[0].ReleaseURL != "https://github.com/nginx/nginx/releases" {
		t.Fatalf("list: %+v", list)
	}
	status := map[string]domain.UpdateStatus{}
	for _, v := range list {
		status[v.Ref] = v.Status
	}
	if status["ghcr.io/me/app:latest"] != domain.UpToDate || status["redis@sha256:pinned"] != domain.UpdatePinned || status["my-build:dev"] != domain.UpdateLocal {
		t.Fatalf("statuses: %+v", status)
	}
	if _, ok := cache.m["gone:1"]; ok {
		t.Fatal("results of removed images must be dropped")
	}
	if len(obs.got) != 2 {
		t.Fatalf("observations: %+v", obs.got)
	}

	// Within 6 h nothing is asked again …
	clk.t = clk.t.Add(time.Hour)
	_ = uc.Check(ctx)
	if reg.calls != 2 {
		t.Fatalf("rate limit: %d calls", reg.calls)
	}
	// … unless the local image changed (pulled).
	images.list[0].RepoDigests = []string{"nginx@sha256:new"}
	_ = uc.Check(ctx)
	if reg.calls != 3 || cache.m["nginx:1.27.1"].Status != domain.UpToDate {
		t.Fatalf("after pull: %d %+v", reg.calls, cache.m["nginx:1.27.1"])
	}
}

func TestUpdateSettingsKeepPasswords(t *testing.T) {
	ctx := context.Background()
	settings := &memSettings{s: UpdateSettings{Credentials: []RegistryCredential{{Host: "ghcr.io", Username: "me", Password: "secret"}}}}
	uc := NewImageUpdates(UpdateDeps{Settings: settings, Clock: &movingClock{}})
	v, err := uc.SaveSettings(ctx, UpdateSettingsInput{Enabled: true, Credentials: []RegistryCredential{{Host: "ghcr.io", Username: "me2"}}})
	if err != nil || settings.s.Credentials[0].Password != "secret" || !v.Registries[0].HasPassword {
		t.Fatalf("%+v %v", settings.s, err)
	}
	for _, bad := range [][]RegistryCredential{
		{{Host: "quay.io", Username: "x"}},                                                           // new without password
		{{Host: "", Username: "x", Password: "p"}},                                                   // no host
		{{Host: "a.io", Username: "x", Password: "p"}, {Host: "a.io", Username: "y", Password: "q"}}, // duplicate
	} {
		if _, err := uc.SaveSettings(ctx, UpdateSettingsInput{Credentials: bad}); !errors.Is(err, kernel.ErrInvalid) {
			t.Fatalf("%+v: %v", bad, err)
		}
	}
}
