package wireguard

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/wireguard/adapter/controller"
	"trisse-home/internal/modules/wireguard/adapter/gateway"
	"trisse-home/internal/modules/wireguard/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the WireGuard rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the WireGuard module; dir is the export directory.
func New(db *sql.DB, dir string, clock kernel.Clock) *Module {
	uc := usecase.NewWireGuard(gateway.NewFileSource(dir), gateway.NewNameStore(db), clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "wireguard" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"wireguard"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/wireguard/peers", m.handlers.List)
	r.Protected("PUT /api/wireguard/names", m.handlers.Rename)
}
