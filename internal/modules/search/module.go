package search

import (
	"context"
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
	modules.Transferable
	handlers *controller.Handlers
	seed     usecase.SeedSampleData
}

// New builds the search module.
func New(db *sql.DB) *Module {
	repo := gateway.NewRepository(db)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(usecase.NewSearcher(repo)), seed: usecase.NewSeedSampleData(repo)}
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

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
