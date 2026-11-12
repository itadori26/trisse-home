package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeContainers struct {
	list []domain.Container
	err  error
}

func (f *fakeContainers) Containers(context.Context) ([]domain.Container, error) {
	return f.list, f.err
}

type memDiscovery struct {
	settings *domain.DiscoverySettings
	last     SyncResult
}

func (m *memDiscovery) DiscoverySettings(context.Context) (domain.DiscoverySettings, error) {
	if m.settings == nil {
		return domain.DefaultDiscoverySettings(), nil
	}
	return *m.settings, nil
}
func (m *memDiscovery) SaveDiscoverySettings(_ context.Context, s domain.DiscoverySettings) error {
	m.settings = &s
	return nil
}
func (m *memDiscovery) LastSync(context.Context) (SyncResult, error)       { return m.last, nil }
func (m *memDiscovery) SaveLastSync(_ context.Context, r SyncResult) error { m.last = r; return nil }

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

func labelled(name, url string) domain.Container {
	return domain.Container{Name: name, Labels: map[string]string{"trisse-home.url": url}}
}

func newDiscoveryFixture(source ContainerSource) (Discovery, *memRepo, *memDiscovery, *movingClock) {
	repo, store := &memRepo{}, &memDiscovery{}
	c := &movingClock{now: time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)}
	deps := Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}}, Checker: fakeChecker{}, History: &memHistory{},
		Observations: kernel.DiscardObservations{}, Clock: c, IDs: &ids{}}
	return NewDiscovery(deps, source, store), repo, store, c
}

func TestDiscoverySync(t *testing.T) {
	ctx := context.Background()
	src := &fakeContainers{list: []domain.Container{
		labelled("jellyfin", "https://jf.lan"), labelled("bad", "not a url"), {Name: "db"},
	}}
	uc, repo, store, _ := newDiscoveryFixture(src)
	repo.items = []domain.Service{{ID: "m", Name: "Router", URL: "http://fritz.box", CheckURL: "http://fritz.box"}}

	v, err := uc.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Available || v.Last.Created != 1 || v.Last.Skipped != 1 || v.Last.Error != "" {
		t.Fatalf("view: %+v", v)
	}
	jf := repo.items[1]
	if jf.Container != "jellyfin" || jf.Position != 1 || jf.Name != "jellyfin" {
		t.Fatalf("created tile: %+v", jf)
	}

	src.list = nil
	if v, _ = uc.Sync(ctx); v.Last.Hidden != 1 || !repo.items[1].Hidden || repo.items[0].Hidden {
		t.Fatalf("vanished container hides only its tile: %+v %+v", v.Last, repo.items)
	}

	src.err = errors.New("dial tcp: refused")
	v, err = uc.Sync(ctx)
	if err != nil || v.Last.Error != errDockerUnreachable || len(repo.items) != 2 {
		t.Fatalf("docker errors are shown, not returned: %+v %v", v, err)
	}
	if store.last.Error == "" {
		t.Fatal("error not stored")
	}
}

func TestDiscoveryEditedTile(t *testing.T) {
	ctx := context.Background()
	src := &fakeContainers{list: []domain.Container{labelled("grafana", "https://g.lan")}}
	uc, repo, _, _ := newDiscoveryFixture(src)
	if _, err := uc.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	services := NewServices(Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}}, Checker: fakeChecker{}, History: &memHistory{},
		Observations: kernel.DiscardObservations{}, Clock: clk{}, IDs: &ids{}})
	if _, err := services.Update(ctx, repo.items[0].ID, ServiceInput{Name: "Dashboards", URL: "https://g.lan"}); err != nil {
		t.Fatal(err)
	}
	if s := repo.items[0]; !s.Edited || s.Container != "grafana" {
		t.Fatalf("a manual edit keeps the container link: %+v", s)
	}
	if _, err := uc.Save(ctx, DiscoveryInput{IntervalMinutes: 5, Rule: string(domain.ManualWins)}); err != nil {
		t.Fatal(err)
	}
	_, _ = uc.Sync(ctx)
	if repo.items[0].Name != "Dashboards" {
		t.Fatal("manual wins keeps the edit")
	}
	_, _ = uc.Save(ctx, DiscoveryInput{IntervalMinutes: 5, Rule: string(domain.LabelsWin)})
	_, _ = uc.Sync(ctx)
	if repo.items[0].Name != "grafana" || repo.items[0].Edited {
		t.Fatalf("labels win overwrites the edit: %+v", repo.items[0])
	}
}

func TestDiscoverySyncIfDue(t *testing.T) {
	ctx := context.Background()
	src := &fakeContainers{list: []domain.Container{labelled("a", "https://a.lan")}}
	uc, repo, _, c := newDiscoveryFixture(src)
	if err := uc.SyncIfDue(ctx); err != nil || len(repo.items) != 0 {
		t.Fatal("disabled discovery must not sync")
	}
	if _, err := uc.Save(ctx, DiscoveryInput{Enabled: true, IntervalMinutes: 10, Rule: "labels"}); err != nil {
		t.Fatal(err)
	}
	if err := uc.SyncIfDue(ctx); err != nil || len(repo.items) != 1 {
		t.Fatalf("first sync: %v %d", err, len(repo.items))
	}
	src.list = append(src.list, labelled("b", "https://b.lan"))
	c.now = c.now.Add(9 * time.Minute)
	_ = uc.SyncIfDue(ctx)
	if len(repo.items) != 1 {
		t.Fatal("synced before the interval")
	}
	c.now = c.now.Add(time.Minute)
	_ = uc.SyncIfDue(ctx)
	if len(repo.items) != 2 {
		t.Fatal("not synced after the interval")
	}
	src.err = errors.New("down")
	c.now = c.now.Add(time.Hour)
	if err := uc.SyncIfDue(ctx); err == nil {
		t.Fatal("the job reports Docker errors for the log")
	}
}

func TestDiscoveryUnavailable(t *testing.T) {
	ctx := context.Background()
	uc, _, _, _ := newDiscoveryFixture(nil)
	if v, err := uc.Get(ctx); err != nil || v.Available {
		t.Fatalf("%+v %v", v, err)
	}
	if _, err := uc.Sync(ctx); !errors.Is(err, domain.ErrDiscoveryUnavailable) {
		t.Fatalf("sync: %v", err)
	}
	if _, err := uc.Save(ctx, DiscoveryInput{Enabled: true, IntervalMinutes: 5, Rule: "labels"}); !errors.Is(err, domain.ErrDiscoveryUnavailable) {
		t.Fatalf("enable: %v", err)
	}
	if _, err := uc.Save(ctx, DiscoveryInput{IntervalMinutes: 0, Rule: "labels"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("validation: %v", err)
	}
	if err := uc.SyncIfDue(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAllSkipsHidden(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{items: []domain.Service{{ID: "1", Name: "Gone", URL: "https://down.example", CheckURL: "https://down.example", Container: "x", Hidden: true}}}
	status := &memStatus{m: map[string]domain.Status{}}
	uc := NewServices(Deps{Repo: repo, Status: status, Checker: fakeChecker{}, History: &memHistory{},
		Observations: kernel.DiscardObservations{}, Clock: clk{}, IDs: &ids{}})
	if err := uc.CheckAll(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := status.Get("1"); ok {
		t.Fatal("hidden tiles are not checked")
	}
	list, _ := uc.List(ctx)
	if !list[0].Hidden || list[0].Container != "x" {
		t.Fatalf("view: %+v", list[0])
	}
}
