package study

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/study/adapter/controller"
	"trisse-home/internal/modules/study/adapter/gateway"
	"trisse-home/internal/modules/study/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the study rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the module (#57); it makes no outbound requests.
func New(db *sql.DB, ids kernel.IDGenerator) *Module {
	uc := usecase.New(gateway.NewRepository(db), ids)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "study" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"study"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/study/modules", m.handlers.List)
	r.Protected("POST /api/study/modules", m.handlers.Create)
	r.Protected("PUT /api/study/modules/{id}", m.handlers.Update)
	r.Protected("DELETE /api/study/modules/{id}", m.handlers.Delete)
	r.Protected("GET /api/study/progress", m.handlers.Progress)
	r.Protected("GET /api/study/settings", m.handlers.Settings)
	r.Protected("PUT /api/study/settings", m.handlers.SaveSettings)
}
