package proxmox

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/proxmox/adapter/controller"
	"trisse-home/internal/modules/proxmox/adapter/gateway"
	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/modules/proxmox/usecase"
	"trisse-home/internal/shared/kernel"
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
}

// Module wires the Proxmox rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Proxmox
	handlers *controller.Handlers
}

// New builds the Proxmox module (#84).
func New(cfg Config) *Module {
	uc := usecase.NewProxmox(usecase.Deps{
		Repo: gateway.NewSettingsRepository(cfg.DB, cfg.Cipher), Client: gateway.NewAPI(cfg.Client, cfg.Pinned),
		Clock: cfg.Clock, Policy: domain.ActionPolicy{Allowed: cfg.ActionGuests}, Audit: cfg.Audit,
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "proxmox" }

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
