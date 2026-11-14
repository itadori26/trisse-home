package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ m map[string]domain.Endpoint }

func (r *memRepo) List(context.Context) ([]domain.Endpoint, error) {
	out := []domain.Endpoint{}
	for _, e := range r.m {
		out = append(out, e)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Endpoint, error) {
	e, ok := r.m[id]
	if !ok {
		return domain.Endpoint{}, domain.ErrNotFound
	}
	return e, nil
}
func (r *memRepo) Save(_ context.Context, e domain.Endpoint) error { r.m[e.ID] = e; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.m[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.m, id)
	return nil
}

type fakeFetcher struct {
	calls int
	doc   any
	err   error
	last  domain.Endpoint
}

func (f *fakeFetcher) Fetch(_ context.Context, e domain.Endpoint) (Response, error) {
	f.calls++
	f.last = e
	return Response{Doc: f.doc, Raw: []byte(`{"temp":21}`)}, f.err
}

// mapPaths evaluates "key" paths against a flat map.
type mapPaths struct{}

func (mapPaths) Validate(p string) error {
	if p == "bad" {
		return errors.New("syntax")
	}
	return nil
}
func (mapPaths) Evaluate(doc any, p string) (any, error) {
	v, ok := doc.(map[string]any)[p]
	if !ok {
		return nil, errors.New(p + " not found")
	}
	return v, nil
}

type clk struct{ now time.Time }

func (c *clk) Now() time.Time { return c.now }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "id" + strconv.Itoa(i.n) }

func setup() (*apiWidget, *memRepo, *fakeFetcher, *clk) {
	repo, f, c := &memRepo{m: map[string]domain.Endpoint{}}, &fakeFetcher{doc: map[string]any{"temp": 21.0}}, &clk{now: time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)}
	return New(repo, f, mapPaths{}, c, &ids{}).(*apiWidget), repo, f, c
}

func input() domain.EndpointInput {
	return domain.EndpointInput{Name: "NAS", URL: "http://nas.lan/api", HeaderName: "X-Api-Key", HeaderValue: "s3cret",
		Interval: time.Minute, Extractions: []domain.Extraction{{Label: "Temp", Path: "temp"}, {Label: "Fan", Path: "fan"}}}
}

func TestEndpointsAndCache(t *testing.T) {
	ctx := context.Background()
	uc, repo, f, c := setup()
	e, err := uc.CreateEndpoint(ctx, input())
	if err != nil || e.ID != "id1" || repo.m["id1"].HeaderValue != "s3cret" {
		t.Fatalf("%+v %v", e, err)
	}
	v, err := uc.FetchValues(ctx, "id1")
	if err != nil || len(v.Values) != 2 || v.Values[0].Text != "21" || v.Values[1].Error == "" || f.last.HeaderValue != "s3cret" {
		t.Fatalf("%+v %v", v, err)
	}
	// Within the interval the cached values are served.
	c.now = c.now.Add(59 * time.Second)
	_, _ = uc.FetchValues(ctx, "id1")
	if f.calls != 1 {
		t.Fatalf("calls %d", f.calls)
	}
	c.now = c.now.Add(2 * time.Second)
	_, _ = uc.FetchValues(ctx, "id1")
	if f.calls != 2 {
		t.Fatalf("after the interval: calls %d", f.calls)
	}
	// Updating forgets the cache; an empty header value with keepHeader keeps the secret.
	in := input()
	in.HeaderValue, in.KeepHeader, in.Name = "", true, "NAS 2"
	if e, err := uc.UpdateEndpoint(ctx, "id1", in); err != nil || e.HeaderValue != "s3cret" {
		t.Fatalf("update: %+v %v", e, err)
	}
	_, _ = uc.FetchValues(ctx, "id1")
	if f.calls != 3 {
		t.Fatalf("update must clear the cache: calls %d", f.calls)
	}
	if err := uc.DeleteEndpoint(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.FetchValues(ctx, "id1"); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("deleted: %v", err)
	}
	if _, err := uc.UpdateEndpoint(ctx, "nope", input()); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	bad := input()
	bad.Extractions[0].Path = "bad"
	if _, err := uc.CreateEndpoint(ctx, bad); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid path: %v", err)
	}
}

func TestErrorsAreShownAndRetriedSooner(t *testing.T) {
	ctx := context.Background()
	uc, _, f, c := setup()
	in := input()
	in.Interval = time.Hour
	e, _ := uc.CreateEndpoint(ctx, in)
	f.err = errors.New("dial tcp: connection refused")
	v, err := uc.FetchValues(ctx, e.ID)
	if err != nil || v.Error == nil || len(v.Values) != 0 {
		t.Fatalf("%+v %v", v, err)
	}
	// Raw error text is not shown; a generic translatable message is.
	if v.Error.Message == "dial tcp: connection refused" || v.Error.Cause == nil {
		t.Fatalf("raw error leaked: %+v", v.Error)
	}
	f.err = nil
	c.now = c.now.Add(61 * time.Second)
	if v, _ := uc.FetchValues(ctx, e.ID); v.Error != nil || f.calls != 2 {
		t.Fatalf("a failed request is retried after a minute, not an hour: %+v calls=%d", v, f.calls)
	}
}

func TestPreviewDoesNotStore(t *testing.T) {
	ctx := context.Background()
	uc, repo, f, _ := setup()
	p, err := uc.Preview(ctx, "", input())
	if err != nil || p.Body != `{"temp":21}` || len(p.Values.Values) != 2 || len(repo.m) != 0 {
		t.Fatalf("%+v %v", p, err)
	}
	e, _ := uc.CreateEndpoint(ctx, input())
	in := input()
	in.HeaderValue, in.KeepHeader = "", true
	if _, err := uc.Preview(ctx, e.ID, in); err != nil || f.last.HeaderValue != "s3cret" {
		t.Fatalf("preview with stored secret: %v %q", err, f.last.HeaderValue)
	}
	if _, err := uc.Preview(ctx, "nope", in); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestEndpointLimit(t *testing.T) {
	ctx := context.Background()
	uc, _, _, _ := setup()
	for i := 0; i < domain.MaxEndpoints; i++ {
		if _, err := uc.CreateEndpoint(ctx, input()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := uc.CreateEndpoint(ctx, input()); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("limit: %v", err)
	}
}
