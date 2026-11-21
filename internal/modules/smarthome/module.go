package smarthome

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"

	"waylight/internal/modules"
	"waylight/internal/modules/smarthome/adapter/controller"
	"waylight/internal/modules/smarthome/adapter/gateway"
	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/modules/smarthome/usecase"
	"waylight/internal/shared/kernel"
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
	// SignIns hands over sign-ins of the setup assistant (#654); may be nil.
	SignIns kernel.SignIns
	// Connections resolves the named connection a widget chose (#610); may
	// be nil.
	Connections modules.ConnectionResolver
}

// Module wires the smart home rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.SmartHome
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
	// repo reads the settings the integrations store copies (#610).
	repo usecase.SettingsRepository
}

// New builds the smart home module (#49) with Home Assistant as provider.
func New(cfg Config) *Module {
	hass := gateway.NewHomeAssistant(cfg.Client)
	repo := gateway.NewSettingsRepository(cfg.DB, cfg.Cipher)
	uc := usecase.NewSmartHome(usecase.Deps{
		Repo: repo, Provider: hass, Conns: connectionsOf(cfg.Connections),
		Clock: cfg.Clock, Policy: domain.ActionPolicy{Allowed: cfg.ActionEntities}, Audit: cfg.Audit,
		Renewer: newRenewer(cfg.Client, cfg.Clock.Now), SignIns: cfg.SignIns,
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), repo: repo, specs: connectionSpecs(hass)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "smarthome" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

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

// CompactStateFor implements modules.CompactParamProvider (#614): the
// device of a smart home tile with one entity, chosen with ?entity=, read
// through the connection in ?connection= (#610), with a toggle or press as
// quick action while actions are enabled for it.
func (m *Module) CompactStateFor(ctx context.Context, params url.Values) (kernel.CompactState, error) {
	conn := params.Get("connection")
	state, e, err := m.uc.Compact(ctx, conn, params.Get("entity"))
	if err != nil {
		return kernel.CompactState{}, err
	}
	state.Action = compactAction(e, conn)
	return state, nil
}

// compactAction is the request behind the entity's action, labelled with
// what it does in the current state; nil when the entity has no action.
// The action goes to the same connection the state was read from.
func compactAction(e usecase.EntityView, connectionID string) *kernel.CompactAction {
	var label kernel.Text
	switch e.Action {
	case domain.ActionToggle:
		label = kernel.NewText("Turn on")
		if e.State == "on" {
			label = kernel.NewText("Turn off")
		}
	case domain.ActionPress:
		label = kernel.NewText("Run")
	default:
		return nil
	}
	// Entity IDs match [a-z0-9_.], so they need no escaping in a path.
	path := "/api/smarthome/entities/" + e.ID + "/action"
	if connectionID != "" {
		path += "?connection=" + url.QueryEscape(connectionID)
	}
	return &kernel.CompactAction{ID: "smarthome." + string(e.Action), Label: label, Method: http.MethodPost, Path: path}
}

// CompactParams implements modules.CompactParamProvider.
func (m *Module) CompactParams() []string { return []string{"entity", "connection"} }

// ConnectionStatus implements modules.StatusReporter (#599): the entity
// list the widget shows tells whether Home Assistant answers.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	_, err := m.uc.Entities(ctx, "")
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// Discovery implements modules.Discoverable (#590): the web app manifest of
// Home Assistant names it and needs no token.
func (m *Module) Discovery() modules.DiscoverySpec {
	return modules.DiscoverySpec{
		Service: "Home Assistant", Images: []string{"home-assistant", "homeassistant"},
		Hostnames: []string{"homeassistant.local", "homeassistant"},
		Ports:     []int{8123}, ProbePath: "/manifest.json", ProbeContains: "Home Assistant",
		ServiceTypes: []string{"_home-assistant._tcp"},
	}
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

// SavedConnections implements modules.ConnectionSource (#610).
func (m *Module) SavedConnections(ctx context.Context) ([]modules.SavedConnection, error) {
	return savedConnections(ctx, m.repo)
}
