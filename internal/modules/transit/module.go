package transit

import (
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/transit/adapter/controller"
	"trisse-home/internal/modules/transit/adapter/gateway"
	"trisse-home/internal/modules/transit/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the transit rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the transit module; apiURL points to a MOTIS instance.
func New(client *http.Client, apiURL string, clock kernel.Clock) *Module {
	if apiURL == "" {
		apiURL = gateway.DefaultURL
	}
	return &Module{handlers: controller.NewHandlers(usecase.NewTransit(gateway.NewMOTIS(client, apiURL), clock))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "transit" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"transit"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/transit/stops", m.handlers.Stops)
	r.Protected("GET /api/transit/departures", m.handlers.Departures)
}
