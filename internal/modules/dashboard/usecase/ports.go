package usecase

import (
	"context"

	"trisse-home/internal/modules/dashboard/domain"
	"trisse-home/internal/shared/kernel"
)

// Input ports.

// GetDashboard returns all tabs with their widgets, creating a default tab
// on first use.
type GetDashboard interface {
	Execute(ctx context.Context) (DashboardResponse, error)
}

// ManageTabs creates, renames, deletes and reorders tabs.
type ManageTabs interface {
	Create(ctx context.Context, req CreateTabRequest) (TabResponse, error)
	Rename(ctx context.Context, req RenameTabRequest) (TabResponse, error)
	Delete(ctx context.Context, tabID string) error
	Reorder(ctx context.Context, req ReorderTabsRequest) error
}

// ManageWidgets adds, configures, places and removes widgets.
type ManageWidgets interface {
	Add(ctx context.Context, req AddWidgetRequest) (WidgetResponse, error)
	Configure(ctx context.Context, req ConfigureWidgetRequest) (WidgetResponse, error)
	SaveLayout(ctx context.Context, req SaveLayoutRequest) error
	Delete(ctx context.Context, widgetID string) error
}

// Output ports.

// TabRepository persists tabs.
type TabRepository interface {
	ListTabs(ctx context.Context) ([]domain.Tab, error)
	GetTab(ctx context.Context, id string) (domain.Tab, error)
	CreateTab(ctx context.Context, t domain.Tab) error
	UpdateTab(ctx context.Context, t domain.Tab) error
	// DeleteTab removes the tab and all its widgets.
	DeleteTab(ctx context.Context, id string) error
	SaveTabOrder(ctx context.Context, tabs []domain.Tab) error
}

// WidgetRepository persists widgets.
type WidgetRepository interface {
	ListWidgets(ctx context.Context) ([]domain.Widget, error)
	ListWidgetsByTab(ctx context.Context, tabID string) ([]domain.Widget, error)
	GetWidget(ctx context.Context, id string) (domain.Widget, error)
	CreateWidget(ctx context.Context, w domain.Widget) error
	UpdateWidget(ctx context.Context, w domain.Widget) error
	SaveWidgetRects(ctx context.Context, widgets []domain.Widget) error
	DeleteWidget(ctx context.Context, id string) error
}

// KindCatalog knows which widget kinds are provided by enabled modules.
type KindCatalog interface {
	Available(kind string) bool
}

// Deps bundles the output ports of the dashboard use cases.
type Deps struct {
	Tabs    TabRepository
	Widgets WidgetRepository
	Kinds   KindCatalog
	IDs     kernel.IDGenerator
}
