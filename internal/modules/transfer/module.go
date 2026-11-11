package transfer

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/transfer/adapter/controller"
	"trisse-home/internal/modules/transfer/adapter/gateway"
	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the transfer rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the transfer module. portables is called per request (usually
// Registry.Portables) so modules registered later are included.
func New(portables func() map[string]modules.Portable, clock kernel.Clock) *Module {
	return &Module{handlers: controller.NewHandlers(usecase.NewTransfer(gateway.Sections(portables), clock))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "transfer" }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/transfer/export", m.handlers.Export)
	r.Protected("POST /api/transfer/import", m.handlers.Import)
}
