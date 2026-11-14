package usecase

import (
	"context"
	"fmt"
	"net/netip"
	"slices"
	"time"

	"trisse-home/internal/modules/dashboard/domain"
	"trisse-home/internal/shared/kernel"
)

// DefaultTabName is the name of the tab created on first use.
const DefaultTabName = "Home"

// defaultWidget describes a widget placed on the first tab.
type defaultWidget struct {
	kind       string
	x, y, w, h int
}

// defaultWidgets give a new installation a useful starting point; kinds of
// disabled modules are skipped.
var defaultWidgets = []defaultWidget{
	{kind: "clock", x: 0, y: 0, w: 3, h: 3},
}

type getDashboard struct{ deps Deps }

// NewGetDashboard returns the GetDashboard use case.
func NewGetDashboard(deps Deps) GetDashboard { return &getDashboard{deps: deps} }

// Execute implements GetDashboard.
func (uc *getDashboard) Execute(ctx context.Context) (DashboardResponse, error) {
	return uc.ExecuteFor(ctx, netip.Addr{})
}

// ExecuteFor implements GetDashboard.
func (uc *getDashboard) ExecuteFor(ctx context.Context, client netip.Addr) (DashboardResponse, error) {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return DashboardResponse{}, fmt.Errorf("list tabs: %w", err)
	}
	// Guests only see the tabs the owner shared with them.
	if p := kernel.PrincipalFrom(ctx); p.TabScoped() {
		tabs = slices.DeleteFunc(tabs, func(t domain.Tab) bool { return !p.CanSeeTab(t.ID()) })
	} else if len(tabs) == 0 {
		tab, err := uc.seed(ctx)
		if err != nil {
			return DashboardResponse{}, err
		}
		tabs = []domain.Tab{tab}
	}
	widgets, err := uc.deps.Widgets.ListWidgets(ctx)
	if err != nil {
		return DashboardResponse{}, fmt.Errorf("list widgets: %w", err)
	}
	res := DashboardResponse{Tabs: make([]TabResponse, 0, len(tabs))}
	if uc.deps.Clock != nil {
		loc := uc.deps.Location
		if loc == nil {
			loc = time.Local
		}
		res.DefaultTabID = domain.DefaultTabID(tabs, uc.deps.Clock.Now().In(loc), client)
	}
	for _, t := range tabs {
		res.Tabs = append(res.Tabs, toTabResponse(t, widgets))
	}
	return res, nil
}

func (uc *getDashboard) seed(ctx context.Context) (domain.Tab, error) {
	tab, err := domain.NewTab(uc.deps.IDs.NewID(), DefaultTabName, 0)
	if err != nil {
		return domain.Tab{}, err
	}
	if err := uc.deps.Tabs.CreateTab(ctx, tab); err != nil {
		return domain.Tab{}, fmt.Errorf("create default tab: %w", err)
	}
	for _, d := range defaultWidgets {
		if !uc.deps.Kinds.Available(d.kind) {
			continue
		}
		rect, err := domain.NewRect(d.x, d.y, d.w, d.h)
		if err != nil {
			return domain.Tab{}, err
		}
		w, err := domain.NewWidget(uc.deps.IDs.NewID(), tab.ID(), d.kind, rect, "")
		if err != nil {
			return domain.Tab{}, err
		}
		if err := uc.deps.Widgets.CreateWidget(ctx, w); err != nil {
			return domain.Tab{}, fmt.Errorf("create default widget: %w", err)
		}
	}
	return tab, nil
}
