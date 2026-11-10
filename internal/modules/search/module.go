package search

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/search/adapter/controller"
	"trisse-home/internal/modules/search/adapter/gateway"
	"trisse-home/internal/modules/search/usecase"
)

// Module wires the search rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the search module.
func New(db *sql.DB) *Module {
	return &Module{handlers: controller.NewHandlers(usecase.NewSearcher(gateway.NewRepository(db)))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "search" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"search"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/search", h.Search)
	r.Public("GET /api/search/opensearch.xml", h.OpenSearch)
	r.Protected("GET /api/search/settings", h.Settings)
	r.Protected("PUT /api/search/settings", h.SetDefaultProvider)
	r.Protected("POST /api/search/shortcuts", h.AddShortcut)
	r.Protected("DELETE /api/search/shortcuts/{keyword}", h.DeleteShortcut)
}
