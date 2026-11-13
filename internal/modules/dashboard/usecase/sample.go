package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/dashboard/domain"
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
}

// NewSeedSampleData returns the SeedSampleData use case for the given layout.
func NewSeedSampleData(deps Deps, tabs []SampleTab) SeedSampleData {
	return &seedSampleData{deps: deps, tabs: tabs}
}

// Execute implements SeedSampleData. An existing dashboard is never touched,
// and widgets of disabled modules are skipped.
func (uc *seedSampleData) Execute(ctx context.Context) error {
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
		if err := uc.deps.Tabs.CreateTab(ctx, tab); err != nil {
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
