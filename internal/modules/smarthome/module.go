package smarthome

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/smarthome/adapter/controller"
	"trisse-home/internal/modules/smarthome/adapter/gateway"
	"trisse-home/internal/modules/smarthome/domain"
	"trisse-home/internal/modules/smarthome/usecase"
	"trisse-home/internal/shared/kernel"
)

// Config configures the smart home module.
type Config struct {
	DB     *sql.DB
	Cipher gateway.Cipher
	Client *http.Client
	// ActionEntities may be switched or pressed (SMARTHOME_ACTIONS_ENTITIES);
	// empty disables actions.
	ActionEntities []string
	Clock          kernel.Clock
	Audit          kernel.AuditSink
}

// Module wires the smart home rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the smart home module (#49) with Home Assistant as provider.
func New(cfg Config) *Module {
	uc := usecase.NewSmartHome(usecase.Deps{
		Repo: gateway.NewSettingsRepository(cfg.DB, cfg.Cipher), Provider: gateway.NewHomeAssistant(cfg.Client),
		Clock: cfg.Clock, Policy: domain.ActionPolicy{Allowed: cfg.ActionEntities}, Audit: cfg.Audit,
	})
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "smarthome" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"smarthome"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/smarthome/settings", m.handlers.Settings)
	r.Protected("PUT /api/smarthome/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/smarthome/entities", m.handlers.Entities)
	r.Protected("GET /api/smarthome/states", m.handlers.States)
	r.Protected("POST /api/smarthome/entities/{id}/action", m.handlers.Act)
}
