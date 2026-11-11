package appearance

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/appearance/adapter/controller"
	"trisse-home/internal/modules/appearance/adapter/gateway"
	"trisse-home/internal/modules/appearance/usecase"
)

// Module wires the appearance rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the appearance module.
func New(db *sql.DB) *Module {
	repo := gateway.NewRepository(db)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(usecase.NewGetAppearance(repo), usecase.NewUpdateAppearance(repo))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "appearance" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/appearance", m.handlers.Get)
	r.Protected("PUT /api/appearance", m.handlers.Update)
}
