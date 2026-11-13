package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/smarthome/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type call struct {
	svc domain.Service
	id  string
}

type fakeProvider struct {
	err    error
	states []domain.Entity
	reads  int
	calls  []call
}

func (f *fakeProvider) States(_ context.Context, s domain.Settings) ([]domain.Entity, error) {
	if f.err != nil {
		return nil, f.err
	}
	if s.Token != "token" {
		return nil, errors.New("the access token was rejected")
	}
	f.reads++
	return f.states, nil
}

func (f *fakeProvider) Call(_ context.Context, _ domain.Settings, svc domain.Service, id string) error {
	f.calls = append(f.calls, call{svc, id})
	return nil
}

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

type audit struct{ events []kernel.AuditEvent }

func (a *audit) Audit(_ context.Context, e kernel.AuditEvent) { a.events = append(a.events, e) }

func setup() (SmartHome, *memRepo, *fakeProvider, *movingClock, *audit) {
	repo := &memRepo{}
	p := &fakeProvider{states: []domain.Entity{
		{ID: "switch.coffee", Name: "Coffee", State: "off"},
		{ID: "sensor.power", Name: "Power", State: "42", Unit: "W"},
		{ID: "switch.heater", Name: "heater", State: "on"},
	}}
	c := &movingClock{now: time.Unix(0, 0)}
	a := &audit{}
	uc := NewSmartHome(Deps{Repo: repo, Provider: p, Clock: c, Policy: domain.ActionPolicy{Allowed: []string{"switch.coffee", "sensor.power"}}, Audit: a})
	return uc, repo, p, c, a
}

func TestSettings(t *testing.T) {
	ctx := context.Background()
	uc, repo, _, _, _ := setup()
	if _, err := uc.States(ctx, []string{"switch.coffee"}); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ha:8123", Token: "wrong"}); !errors.Is(err, kernel.ErrInvalid) || repo.s != nil {
		t.Fatalf("a failing connection test must not save: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ha:8123", Token: "token"}); err != nil {
		t.Fatal(err)
	}
	v, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ha:8123"})
	if err != nil || !v.HasToken || repo.s.Token != "token" {
		t.Fatalf("an empty token keeps the stored one: %+v %v", v, err)
	}
}

func TestStatesAndCache(t *testing.T) {
	ctx := context.Background()
	uc, repo, p, c, _ := setup()
	repo.s = &domain.Settings{URL: "http://ha", Token: "token"}

	list, err := uc.States(ctx, []string{"sensor.power", "switch.unknown", "switch.coffee", "switch.heater"})
	if err != nil || len(list) != 3 || list[0].ID != "sensor.power" || list[1].ID != "switch.coffee" {
		t.Fatalf("requested order, unknown skipped: %+v %v", list, err)
	}
	if list[0].Action != domain.ActionNone || list[1].Action != domain.ActionToggle || list[2].Action != domain.ActionNone {
		t.Fatalf("actions only for allow-listed switches: %+v", list)
	}
	_, _ = uc.States(ctx, []string{"switch.coffee"})
	if p.reads != 1 {
		t.Fatalf("states are reused within the TTL, got %d reads", p.reads)
	}
	c.now = c.now.Add(statesTTL)
	_, _ = uc.States(ctx, []string{"switch.coffee"})
	if p.reads != 2 {
		t.Fatalf("states are fetched again after the TTL, got %d reads", p.reads)
	}
	all, _ := uc.Entities(ctx)
	if len(all) != 3 || all[0].Name != "Coffee" || all[2].Name != "Power" {
		t.Fatalf("entities sorted by name: %+v", all)
	}
	many := make([]string, MaxEntitiesPerWidget+1)
	if _, err := uc.States(ctx, many); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("too many entities: %v", err)
	}
	p.err = errors.New("down")
	c.now = c.now.Add(statesTTL)
	if _, err := uc.States(ctx, []string{"switch.coffee"}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("unreachable: %v", err)
	}
}

func TestAct(t *testing.T) {
	ctx := context.Background()
	uc, repo, p, _, a := setup()
	repo.s = &domain.Settings{URL: "http://ha", Token: "token"}
	for id, want := range map[string]error{
		"switch.heater":  domain.ErrActionNotAllowed,
		"sensor.power":   domain.ErrNoAction,
		"switch/../etc":  domain.ErrEntityNotFound,
		"switch.missing": domain.ErrActionNotAllowed,
	} {
		if err := uc.Act(ctx, id); !errors.Is(err, want) {
			t.Errorf("%s: got %v, want %v", id, err, want)
		}
	}
	if err := uc.Act(ctx, "switch.coffee"); err != nil {
		t.Fatal(err)
	}
	if len(p.calls) != 1 || p.calls[0].svc != (domain.Service{Domain: "switch", Name: "toggle"}) || p.calls[0].id != "switch.coffee" {
		t.Fatalf("calls: %+v", p.calls)
	}
	if len(a.events) != 1 || a.events[0].Type != "smarthome_action" {
		t.Fatalf("audit: %+v", a.events)
	}
	reads := p.reads
	_, _ = uc.States(ctx, []string{"switch.coffee"})
	if p.reads != reads+1 {
		t.Fatal("an action invalidates the cached states")
	}
}
