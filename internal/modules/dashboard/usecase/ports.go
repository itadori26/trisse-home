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
	// SetBackground sets the tab's background image ("" = global one).
	SetBackground(ctx context.Context, tabID, background string) (TabResponse, error)
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

// TemplateRepository persists tab templates (output port).
type TemplateRepository interface {
	ListTemplates(ctx context.Context) ([]domain.Template, error)
	GetTemplate(ctx context.Context, id string) (domain.Template, error)
	CreateTemplate(ctx context.Context, t domain.Template) error
	DeleteTemplate(ctx context.Context, id string) error
	// CreateTabWithWidgets stores a new tab and its widgets atomically.
	CreateTabWithWidgets(ctx context.Context, t domain.Tab, widgets []domain.Widget) error
}

// ManageTemplates saves tabs as templates and creates tabs from them.
type ManageTemplates interface {
	List(ctx context.Context) ([]TemplateResponse, error)
	SaveTab(ctx context.Context, tabID, name string) (TemplateResponse, error)
	Delete(ctx context.Context, id string) error
	CreateTab(ctx context.Context, templateID, name string) (TabResponse, error)
}

// KindCatalog knows which widget kinds are provided by enabled modules.
type KindCatalog interface {
	Available(kind string) bool
}

// Deps bundles the output ports of the dashboard use cases.
type Deps struct {
	Tabs      TabRepository
	Widgets   WidgetRepository
	Templates TemplateRepository
	Kinds     KindCatalog
	IDs       kernel.IDGenerator
}
