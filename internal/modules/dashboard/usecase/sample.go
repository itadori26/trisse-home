package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

// SeedSampleData fills an empty dashboard with example tabs and widgets
// (input port).
type SeedSampleData interface {
	Execute(ctx context.Context) error
}

// SampleWidget is a widget of the sample layout.
type SampleWidget struct {
	Kind       string
	X, Y, W, H int
	Config     string
}

// SampleTab is a tab of the sample layout.
type SampleTab struct {
	Name    string
	Widgets []SampleWidget
}

// EverydayLayout is the sample layout of the "everyday" start page (#281):
// only the Home tab.
func EverydayLayout(weatherConfig string) []SampleTab {
	return SampleLayout(weatherConfig)[:1]
}

// SampleLayout returns the layout shown after the first setup. It presents
// every widget kind; weatherConfig is the JSON configuration of the weather
// widget (location chosen by the operator).
func SampleLayout(weatherConfig string) []SampleTab {
	return []SampleTab{
		{Name: "Home", Widgets: []SampleWidget{
			{Kind: "clock", X: 0, Y: 0, W: 3, H: 3},
			{Kind: "weather", X: 3, Y: 0, W: 4, H: 4, Config: weatherConfig},
			{Kind: "todos", X: 7, Y: 0, W: 5, H: 4},
			{Kind: "notes", X: 0, Y: 3, W: 3, H: 6, Config: `{"scope":"pinned"}`},
			{Kind: "bookmarks", X: 3, Y: 4, W: 4, H: 5},
			{Kind: "rss", X: 7, Y: 4, W: 5, H: 6, Config: `{"unreadOnly":true,"limit":15}`},
			{Kind: "calendar", X: 0, Y: 9, W: 3, H: 3, Config: `{"days":7}`},
		}},
		{Name: "Server", Widgets: []SampleWidget{
			{Kind: "services", X: 0, Y: 0, W: 12, H: 2},
			{Kind: "system", X: 0, Y: 2, W: 4, H: 5},
			{Kind: "docker", X: 4, Y: 2, W: 4, H: 7},
			{Kind: "adguard", X: 8, Y: 2, W: 4, H: 4},
		}},
	}
}

type seedSampleData struct {
	deps Deps
	tabs []SampleTab
	// mark flags the tabs as example tabs, which the template choice on
	// first start may replace (#670).
	mark bool
}

// NewSeedSampleData returns the SeedSampleData use case for the given
// layout. Its tabs are marked as example tabs.
func NewSeedSampleData(deps Deps, tabs []SampleTab) SeedSampleData {
	return &seedSampleData{deps: deps, tabs: tabs, mark: true}
}

// NewSeedShowcase returns the SeedSampleData use case for the demo
// showcase, whose tabs are the demo itself and are never replaced.
func NewSeedShowcase(deps Deps, tabs []SampleTab) SeedSampleData {
	return &seedSampleData{deps: deps, tabs: tabs}
}

// Execute implements SeedSampleData. An existing dashboard is never touched,
// and widgets of disabled modules are skipped.
func (uc *seedSampleData) Execute(ctx context.Context) error {
	// Sample tabs are the admin's; seeding runs right after setup as the
	// system, which would otherwise create and count tabs of every account.
	ctx = kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	existing, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return fmt.Errorf("list tabs: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}
	for pos, st := range uc.tabs {
		tab, err := domain.NewTab(uc.deps.IDs.NewID(), st.Name, pos)
		if err != nil {
			return err
		}
		if err := uc.deps.Tabs.CreateTab(ctx, tab.WithSample(uc.mark)); err != nil {
			return fmt.Errorf("create tab: %w", err)
		}
		for _, sw := range st.Widgets {
			if !uc.deps.Kinds.Available(sw.Kind) {
				continue
			}
			rect, err := domain.NewRect(sw.X, sw.Y, sw.W, sw.H)
			if err != nil {
				return err
			}
			w, err := domain.NewWidget(uc.deps.IDs.NewID(), tab.ID(), sw.Kind, rect, sw.Config)
			if err != nil {
				return err
			}
			if err := uc.deps.Widgets.CreateWidget(ctx, w); err != nil {
				return fmt.Errorf("create widget: %w", err)
			}
		}
	}
	return nil
}

// SampleTabs settles the example tabs once the template choice on first
// start is made (#670): replaced by a template or "start empty", they are
// removed; kept, they become ordinary tabs (input port).
type SampleTabs interface {
	// Discard removes the account's example tabs and closes the gaps in the
	// tab order. Without other tabs the next load creates the default tab.
	Discard(ctx context.Context) error
	// Keep turns the account's example tabs into ordinary tabs.
	Keep(ctx context.Context) error
}

type sampleTabs struct{ deps Deps }

// NewSampleTabs returns the SampleTabs use case.
func NewSampleTabs(deps Deps) SampleTabs { return &sampleTabs{deps: deps} }

// ownSample splits the tabs the principal may change into example tabs and
// the rest. Team tabs others shared are left alone either way.
func (uc *sampleTabs) ownSample(ctx context.Context) (sample, rest []domain.Tab, err error) {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list tabs: %w", err)
	}
	p := kernel.PrincipalFrom(ctx)
	for _, t := range tabs {
		mine := t.Access().Owner == "" || t.Access().Mine(p)
		if mine && t.Sample() {
			sample = append(sample, t)
		} else {
			rest = append(rest, t)
		}
	}
	return sample, rest, nil
}

// Discard implements SampleTabs.
func (uc *sampleTabs) Discard(ctx context.Context) error {
	sample, rest, err := uc.ownSample(ctx)
	if err != nil || len(sample) == 0 {
		return err
	}
	for _, t := range sample {
		if err := uc.deps.Tabs.DeleteTab(ctx, t.ID()); err != nil {
			return fmt.Errorf("delete tab: %w", err)
		}
	}
	for i := range rest {
		rest[i].MoveTo(i)
	}
	if err := uc.deps.Tabs.SaveTabOrder(ctx, rest); err != nil {
		return fmt.Errorf("save tab order: %w", err)
	}
	return nil
}

// Keep implements SampleTabs.
func (uc *sampleTabs) Keep(ctx context.Context) error {
	sample, _, err := uc.ownSample(ctx)
	if err != nil {
		return err
	}
	for _, t := range sample {
		if err := uc.deps.Tabs.UpdateTab(ctx, t.WithSample(false)); err != nil {
			return fmt.Errorf("update tab: %w", err)
		}
	}
	return nil
}
