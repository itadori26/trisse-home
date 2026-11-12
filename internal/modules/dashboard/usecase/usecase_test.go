package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/dashboard/domain"
)

type memRepo struct {
	tabs      []domain.Tab
	widgets   map[string]domain.Widget
	templates []domain.Template
}

func (m *memRepo) ListTemplates(context.Context) ([]domain.Template, error) {
	return slices.Clone(m.templates), nil
}

func (m *memRepo) GetTemplate(_ context.Context, id string) (domain.Template, error) {
	for _, t := range m.templates {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.Template{}, domain.ErrTemplateNotFound
}

func (m *memRepo) CreateTemplate(_ context.Context, t domain.Template) error {
	m.templates = append(m.templates, t)
	return nil
}

func (m *memRepo) DeleteTemplate(_ context.Context, id string) error {
	m.templates = slices.DeleteFunc(m.templates, func(t domain.Template) bool { return t.ID == id })
	return nil
}

func (m *memRepo) CreateTabWithWidgets(_ context.Context, t domain.Tab, ws []domain.Widget) error {
	m.tabs = append(m.tabs, t)
	for _, w := range ws {
		m.widgets[w.ID()] = w
	}
	return nil
}

func newMemRepo() *memRepo { return &memRepo{widgets: map[string]domain.Widget{}} }

func (m *memRepo) ListTabs(context.Context) ([]domain.Tab, error) {
	out := slices.Clone(m.tabs)
	slices.SortFunc(out, func(a, b domain.Tab) int { return a.Position() - b.Position() })
	return out, nil
}

func (m *memRepo) GetTab(_ context.Context, id string) (domain.Tab, error) {
	for _, t := range m.tabs {
		if t.ID() == id {
			return t, nil
		}
	}
	return domain.Tab{}, domain.ErrTabNotFound
}

func (m *memRepo) CreateTab(_ context.Context, t domain.Tab) error {
	m.tabs = append(m.tabs, t)
	return nil
}

func (m *memRepo) UpdateTab(_ context.Context, t domain.Tab) error {
	for i := range m.tabs {
		if m.tabs[i].ID() == t.ID() {
			m.tabs[i] = t
		}
	}
	return nil
}

func (m *memRepo) DeleteTab(_ context.Context, id string) error {
	m.tabs = slices.DeleteFunc(m.tabs, func(t domain.Tab) bool { return t.ID() == id })
	for k, w := range m.widgets {
		if w.TabID() == id {
			delete(m.widgets, k)
		}
	}
	return nil
}

func (m *memRepo) SaveTabOrder(ctx context.Context, tabs []domain.Tab) error {
	for _, t := range tabs {
		_ = m.UpdateTab(ctx, t)
	}
	return nil
}

func (m *memRepo) ListWidgets(context.Context) ([]domain.Widget, error) {
	out := make([]domain.Widget, 0, len(m.widgets))
	for _, w := range m.widgets {
		out = append(out, w)
	}
	return out, nil
}

func (m *memRepo) ListWidgetsByTab(_ context.Context, tabID string) ([]domain.Widget, error) {
	var out []domain.Widget
	for _, w := range m.widgets {
		if w.TabID() == tabID {
			out = append(out, w)
		}
	}
	return out, nil
}

func (m *memRepo) GetWidget(_ context.Context, id string) (domain.Widget, error) {
	w, ok := m.widgets[id]
	if !ok {
		return domain.Widget{}, domain.ErrWidgetNotFound
	}
	return w, nil
}

func (m *memRepo) CreateWidget(_ context.Context, w domain.Widget) error {
	m.widgets[w.ID()] = w
	return nil
}
func (m *memRepo) UpdateWidget(_ context.Context, w domain.Widget) error {
	m.widgets[w.ID()] = w
	return nil
}
func (m *memRepo) DeleteWidget(_ context.Context, id string) error { delete(m.widgets, id); return nil }

func (m *memRepo) SaveWidgetRects(_ context.Context, ws []domain.Widget) error {
	for _, w := range ws {
		m.widgets[w.ID()] = w
	}
	return nil
}

type kinds []string

func (k kinds) Available(kind string) bool { return slices.Contains(k, kind) }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return fmt.Sprintf("id%d", s.n) }

func newDeps(available ...string) (Deps, *memRepo) {
	repo := newMemRepo()
	return Deps{Tabs: repo, Widgets: repo, Templates: repo, Kinds: kinds(available), IDs: &seqIDs{}}, repo
}

var ctx = context.Background()

func TestGetDashboardSeedsDefaultTab(t *testing.T) {
	deps, repo := newDeps("clock")
	res, err := NewGetDashboard(deps).Execute(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tabs) != 1 || res.Tabs[0].Name != DefaultTabName {
		t.Fatalf("unexpected tabs %+v", res.Tabs)
	}
	if len(res.Tabs[0].Widgets) != 1 || res.Tabs[0].Widgets[0].Kind != "clock" {
		t.Fatalf("expected only the available default widget, got %+v", res.Tabs[0].Widgets)
	}
	if _, err := NewGetDashboard(deps).Execute(ctx); err != nil || len(repo.tabs) != 1 {
		t.Fatal("seeding must happen only once")
	}
}

func TestTabsLifecycle(t *testing.T) {
	deps, _ := newDeps()
	tabs := NewManageTabs(deps)
	a, _ := tabs.Create(ctx, CreateTabRequest{Name: "A"})
	b, _ := tabs.Create(ctx, CreateTabRequest{Name: "B"})
	if b.Position != 1 {
		t.Fatalf("new tab not appended: %d", b.Position)
	}
	if err := tabs.Reorder(ctx, ReorderTabsRequest{TabIDs: []string{b.ID}}); !errors.Is(err, domain.ErrTabOrderMismatch) {
		t.Fatalf("expected mismatch, got %v", err)
	}
	if err := tabs.Reorder(ctx, ReorderTabsRequest{TabIDs: []string{b.ID, a.ID}}); err != nil {
		t.Fatal(err)
	}
	dash, _ := NewGetDashboard(deps).Execute(ctx)
	if dash.Tabs[0].ID != b.ID {
		t.Fatal("reorder not applied")
	}
	if _, err := tabs.Rename(ctx, RenameTabRequest{TabID: a.ID, Name: " "}); err == nil {
		t.Fatal("expected validation error")
	}
	if err := tabs.Delete(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if err := tabs.Delete(ctx, b.ID); !errors.Is(err, domain.ErrLastTab) {
		t.Fatalf("expected ErrLastTab, got %v", err)
	}
	if err := tabs.Delete(ctx, "missing"); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestWidgets(t *testing.T) {
	deps, repo := newDeps("weather")
	tab, _ := NewManageTabs(deps).Create(ctx, CreateTabRequest{Name: "A"})
	other, _ := NewManageTabs(deps).Create(ctx, CreateTabRequest{Name: "B"})
	widgets := NewManageWidgets(deps)

	if _, err := widgets.Add(ctx, AddWidgetRequest{TabID: tab.ID, Kind: "docker", W: 2, H: 2}); !errors.Is(err, domain.ErrUnknownWidgetKind) {
		t.Fatalf("expected unknown kind, got %v", err)
	}
	if _, err := widgets.Add(ctx, AddWidgetRequest{TabID: "nope", Kind: "weather", W: 2, H: 2}); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("expected tab not found, got %v", err)
	}
	w, err := widgets.Add(ctx, AddWidgetRequest{TabID: tab.ID, Kind: "weather", W: 3, H: 2, Config: `{"lat":1}`})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := widgets.Configure(ctx, ConfigureWidgetRequest{WidgetID: w.ID, Config: `{"lat":2}`}); err != nil {
		t.Fatal(err)
	}
	err = widgets.SaveLayout(ctx, SaveLayoutRequest{TabID: tab.ID, Placements: []WidgetPlacement{{WidgetID: w.ID, X: 4, Y: 1, W: 4, H: 3}}})
	if err != nil {
		t.Fatal(err)
	}
	if r := repo.widgets[w.ID].Rect(); r.X != 4 || r.H != 3 || repo.widgets[w.ID].Config() != `{"lat":2}` {
		t.Fatalf("layout/config not stored: %+v", r)
	}
	err = widgets.SaveLayout(ctx, SaveLayoutRequest{TabID: other.ID, Placements: []WidgetPlacement{{WidgetID: w.ID, W: 1, H: 1}}})
	if !errors.Is(err, domain.ErrWidgetNotFound) {
		t.Fatalf("moving a foreign widget must fail, got %v", err)
	}
	if err := widgets.Delete(ctx, w.ID); err != nil || len(repo.widgets) != 0 {
		t.Fatalf("delete failed: %v", err)
	}
}

func TestTemplatesAndBackground(t *testing.T) {
	deps, repo := newDeps("clock", "notes")
	tabs, widgets, tpls := NewManageTabs(deps), NewManageWidgets(deps), NewManageTemplates(deps)
	home, _ := tabs.Create(ctx, CreateTabRequest{Name: "Home"})
	_, _ = widgets.Add(ctx, AddWidgetRequest{TabID: home.ID, Kind: "clock", W: 3, H: 2, Config: `{"tz":"UTC"}`})
	_, _ = widgets.Add(ctx, AddWidgetRequest{TabID: home.ID, Kind: "notes", X: 3, W: 3, H: 2})
	tpl, err := tpls.SaveTab(ctx, home.ID, "Desk")
	if err != nil || tpl.Widgets != 2 {
		t.Fatalf("save: %+v %v", tpl, err)
	}
	// A disabled module's widgets are skipped when the template is applied.
	deps.Kinds = kinds{"clock"}
	tab, err := NewManageTemplates(deps).CreateTab(ctx, tpl.ID, "")
	if err != nil || tab.Name != "Desk" || len(tab.Widgets) != 1 || tab.Widgets[0].Config != `{"tz":"UTC"}` || len(repo.tabs) != 2 {
		t.Fatalf("create: %+v %v", tab, err)
	}
	if _, err := tabs.SetBackground(ctx, tab.ID, "ftp://x"); err == nil {
		t.Fatal("invalid background accepted")
	}
	res, err := tabs.SetBackground(ctx, tab.ID, "https://img.example/a.jpg")
	if err != nil || res.Background != "https://img.example/a.jpg" {
		t.Fatalf("background: %+v %v", res, err)
	}
	if err := tpls.Delete(ctx, "nope"); err == nil {
		t.Fatal("unknown template deleted")
	}
}

type fixedClock time.Time

func (c fixedClock) Now() time.Time { return time.Time(c) }

func TestDefaultTabRules(t *testing.T) {
	deps, _ := newDeps()
	tabs := NewManageTabs(deps)
	home, _ := tabs.Create(ctx, CreateTabRequest{Name: "Home"})
	evening, _ := tabs.Create(ctx, CreateTabRequest{Name: "Evening"})
	if _, err := tabs.SetRules(ctx, evening.ID, []domain.RuleInput{{Days: []int{9}}}); err == nil {
		t.Fatal("invalid rule accepted")
	}
	if _, err := tabs.SetRules(ctx, "nope", nil); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("unknown tab: %v", err)
	}
	res, err := tabs.SetRules(ctx, evening.ID, []domain.RuleInput{
		{From: "18:00", To: "01:00"},
		{Networks: []string{"10.13.13.0/24"}},
	})
	if err != nil || len(res.Rules) != 2 || res.Rules[0].From != "18:00" || res.Rules[1].Networks[0] != "10.13.13.0/24" {
		t.Fatalf("set rules: %+v %v", res, err)
	}

	// Thursday 12 November 2026, evaluated in the configured location.
	cet := time.FixedZone("CET", 3600)
	resolve := func(utc time.Time, ip string) string {
		t.Helper()
		d := deps
		d.Clock, d.Location = fixedClock(utc), cet
		var addr netip.Addr
		if ip != "" {
			addr = netip.MustParseAddr(ip)
		}
		got, err := NewGetDashboard(d).ExecuteFor(ctx, addr)
		if err != nil {
			t.Fatal(err)
		}
		return got.DefaultTabID
	}
	if got := resolve(time.Date(2026, 11, 12, 17, 30, 0, 0, time.UTC), ""); got != evening.ID {
		t.Fatalf("18:30 local must open the evening tab, got %q", got)
	}
	if got := resolve(time.Date(2026, 11, 12, 16, 30, 0, 0, time.UTC), ""); got != "" {
		t.Fatalf("17:30 local has no matching rule, got %q", got)
	}
	if got := resolve(time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC), "10.13.13.5"); got != evening.ID {
		t.Fatalf("VPN client, got %q", got)
	}
	// Order resolves conflicts: a matching rule on the first tab wins.
	if _, err := tabs.SetRules(ctx, home.ID, []domain.RuleInput{{}}); err != nil {
		t.Fatal(err)
	}
	if got := resolve(time.Date(2026, 11, 12, 17, 30, 0, 0, time.UTC), ""); got != home.ID {
		t.Fatalf("first tab must win, got %q", got)
	}
	// Without a clock no default tab is resolved.
	if got, _ := NewGetDashboard(deps).Execute(ctx); got.DefaultTabID != "" {
		t.Fatalf("no clock, got %q", got.DefaultTabID)
	}
}
