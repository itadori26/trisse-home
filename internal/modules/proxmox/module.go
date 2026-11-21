package proxmox

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/proxmox/adapter/controller"
	"waylight/internal/modules/proxmox/adapter/gateway"
	"waylight/internal/modules/proxmox/domain"
	"waylight/internal/modules/proxmox/usecase"
	"waylight/internal/shared/kernel"
)

const refreshInterval = 30 * time.Second

// Config configures the Proxmox module.
type Config struct {
	DB     *sql.DB
	Cipher gateway.Cipher
	Client *http.Client
	// Pinned builds clients for a pinned certificate fingerprint; may be nil.
	Pinned gateway.PinnedClientFactory
	// ActionGuests are the guest IDs that may be started and shut down
	// (PROXMOX_ACTIONS_GUESTS); empty disables actions.
	ActionGuests []int
	Clock        kernel.Clock
	Audit        kernel.AuditSink
	// Connections resolves the named connection a widget chose (#610); may
	// be nil.
	Connections modules.ConnectionResolver
}

// Module wires the Proxmox rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Proxmox
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
	// repo reads the settings the integrations store copies (#610).
	repo usecase.SettingsRepository
}

// New builds the Proxmox module (#84).
func New(cfg Config) *Module {
	api := gateway.NewAPI(cfg.Client, cfg.Pinned)
	repo := gateway.NewSettingsRepository(cfg.DB, cfg.Cipher)
	uc := usecase.NewProxmox(usecase.Deps{
		Repo: repo, Client: api, Conns: connectionsOf(cfg.Connections),
		Clock: cfg.Clock, Policy: domain.ActionPolicy{Allowed: cfg.ActionGuests}, Audit: cfg.Audit,
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), repo: repo, specs: connectionSpecs(api)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "proxmox" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"proxmox"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/proxmox/settings", m.handlers.Settings)
	r.Protected("PUT /api/proxmox/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/proxmox/status", m.handlers.Status)
	r.Protected("POST /api/proxmox/guests/{id}/{action}", m.handlers.Act)
}

// Jobs implements modules.Module; the job does nothing until configured.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "proxmox.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

// ConnectionStatus implements modules.StatusReporter (#592): the status the
// job keeps says whether the last refresh reached Proxmox.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	v, err := m.uc.Status(ctx, "")
	if err == nil && v.Error != "" {
		err = errUnreachable
	}
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// CompactStateFor implements modules.CompactParamProvider (#641): running
// guests and whether a node or the API needs a look, read through the
// connection in ?connection= (#610).
func (m *Module) CompactStateFor(ctx context.Context, params url.Values) (kernel.CompactState, error) {
	v, err := m.uc.Status(ctx, params.Get("connection"))
	if err != nil {
		return kernel.CompactState{}, err
	}
	return usecase.CompactState(v), nil
}

// CompactParams implements modules.CompactParamProvider.
func (m *Module) CompactParams() []string { return []string{"connection"} }

var errUnreachable = kernel.NewPublicError(kernel.ErrUnavailable, "Proxmox VE is not reachable.")

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

// SavedConnections implements modules.ConnectionSource (#610).
func (m *Module) SavedConnections(ctx context.Context) ([]modules.SavedConnection, error) {
	return savedConnections(ctx, m.repo)
}
