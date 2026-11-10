package dashboard

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/dashboard/adapter/controller"
	"trisse-home/internal/modules/dashboard/adapter/gateway"
	"trisse-home/internal/modules/dashboard/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the dashboard rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the dashboard module. widgetKinds lists the widget kinds of all
// enabled modules.
func New(db *sql.DB, ids kernel.IDGenerator, widgetKinds []string) *Module {
	repo := gateway.NewRepository(db)
	deps := usecase.Deps{Tabs: repo, Widgets: repo, Kinds: gateway.NewKindCatalog(widgetKinds), IDs: ids}
	return &Module{handlers: controller.NewHandlers(
		usecase.NewGetDashboard(deps), usecase.NewManageTabs(deps), usecase.NewManageWidgets(deps),
	)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "dashboard" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/dashboard", h.Get)
	r.Protected("POST /api/dashboard/tabs", h.CreateTab)
	r.Protected("PUT /api/dashboard/tabs/order", h.ReorderTabs)
	r.Protected("PATCH /api/dashboard/tabs/{id}", h.RenameTab)
	r.Protected("DELETE /api/dashboard/tabs/{id}", h.DeleteTab)
	r.Protected("POST /api/dashboard/tabs/{id}/widgets", h.AddWidget)
	r.Protected("PUT /api/dashboard/tabs/{id}/layout", h.SaveLayout)
	r.Protected("PATCH /api/dashboard/widgets/{id}", h.ConfigureWidget)
	r.Protected("DELETE /api/dashboard/widgets/{id}", h.DeleteWidget)
}
