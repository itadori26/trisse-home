package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/arr/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ list []domain.Instance }

func (m *memRepo) List(context.Context) ([]domain.Instance, error) { return slices.Clone(m.list), nil }
func (m *memRepo) Get(_ context.Context, id string) (domain.Instance, error) {
	for _, in := range m.list {
		if in.ID == id {
			return in, nil
		}
	}
	return domain.Instance{}, domain.ErrInstanceNotFound
}
func (m *memRepo) Save(_ context.Context, in domain.Instance) error {
	for i := range m.list {
		if m.list[i].ID == in.ID {
			m.list[i] = in
			return nil
		}
	}
	m.list = append(m.list, in)
	return nil
}
func (m *memRepo) Delete(_ context.Context, id string) error {
	m.list = slices.DeleteFunc(m.list, func(in domain.Instance) bool { return in.ID == id })
	return nil
}

type fakeApp struct {
	down  bool
	calls int
}

func (f *fakeApp) Check(_ context.Context, in domain.Instance) error {
	if in.APIKey != "key" {
		return errors.New("the API key was rejected")
	}
	return nil
}

func (f *fakeApp) Snapshot(_ context.Context, in domain.Instance, from, _ time.Time) (domain.Snapshot, error) {
	f.calls++
	if f.down {
		return domain.Snapshot{}, errors.New("timeout")
	}
	return domain.Snapshot{Upcoming: []domain.Release{{Instance: in.Name, Title: in.Name, Date: from.Add(time.Duration(len(in.Name)) * time.Hour)}}}, nil
}

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint(i.n) }

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

func setup() (Arr, *memRepo, *fakeApp, *fakeApp, *movingClock) {
	repo, sonarr, radarr := &memRepo{}, &fakeApp{}, &fakeApp{}
	c := &movingClock{now: time.Unix(0, 0)}
	uc := NewArr(Deps{Repo: repo, Providers: map[domain.App]Provider{domain.AppSonarr: sonarr, domain.AppRadarr: radarr, domain.AppProwlarr: &fakeApp{}},
		Clock: c, IDs: &ids{}})
	return uc, repo, sonarr, radarr, c
}

func TestInstances(t *testing.T) {
	ctx := context.Background()
	uc, repo, _, _, _ := setup()
	if _, err := uc.AddInstance(ctx, InstanceInput{App: "sonarr", URL: "http://s", APIKey: "bad"}); !errors.Is(err, kernel.ErrInvalid) || len(repo.list) != 0 {
		t.Fatalf("a failing check must not save: %v", err)
	}
	v, err := uc.AddInstance(ctx, InstanceInput{App: "sonarr", URL: "http://s", APIKey: "key"})
	if err != nil || v.Name != "Sonarr" {
		t.Fatalf("add: %+v %v", v, err)
	}
	if _, err := uc.UpdateInstance(ctx, v.ID, InstanceInput{App: "sonarr", Name: "TV", URL: "http://s2"}); err != nil || repo.list[0].APIKey != "key" {
		t.Fatalf("an empty key keeps the stored one: %v", err)
	}
	list, _ := uc.Instances(ctx)
	if len(list) != 1 || list[0].Name != "TV" {
		t.Fatalf("list: %+v", list)
	}
	if err := uc.DeleteInstance(ctx, "nope"); !errors.Is(err, domain.ErrInstanceNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	for range domain.MaxInstances {
		repo.list = append(repo.list, domain.Instance{ID: "x"})
	}
	if _, err := uc.AddInstance(ctx, InstanceInput{App: "radarr", URL: "http://r", APIKey: "key"}); !errors.Is(err, domain.ErrTooManyInstances) {
		t.Fatalf("limit: %v", err)
	}
}

func TestOverview(t *testing.T) {
	ctx := context.Background()
	uc, repo, sonarr, radarr, c := setup()
	if _, err := uc.Overview(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
	repo.list = []domain.Instance{{ID: "1", App: domain.AppSonarr, Name: "Sonarr4K"}, {ID: "2", App: domain.AppRadarr, Name: "Radarr"}}
	radarr.down = true
	o, err := uc.Overview(ctx)
	if err != nil || len(o.Upcoming) != 1 || len(o.Errors) != 1 || o.Errors[0] != "Radarr (timeout)" {
		t.Fatalf("a failing instance is reported, the others shown: %+v %v", o, err)
	}
	_, _ = uc.Overview(ctx)
	if sonarr.calls != 1 {
		t.Fatalf("cached within the TTL, got %d calls", sonarr.calls)
	}
	radarr.down = false
	c.now = c.now.Add(overviewTTL)
	o, _ = uc.Overview(ctx)
	if len(o.Upcoming) != 2 || o.Upcoming[0].Title != "Radarr" {
		t.Fatalf("merged and sorted after the TTL: %+v", o.Upcoming)
	}
	if err := uc.DeleteInstance(ctx, "2"); err != nil {
		t.Fatal(err)
	}
	if o, _ := uc.Overview(ctx); len(o.Upcoming) != 1 {
		t.Fatalf("deleting invalidates the cache: %+v", o.Upcoming)
	}
}
