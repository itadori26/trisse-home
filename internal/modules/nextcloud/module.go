package nextcloud

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules/nextcloud/domain"

	"waylight/internal/modules"
	"waylight/internal/modules/nextcloud/adapter/controller"
	"waylight/internal/modules/nextcloud/adapter/gateway"
	"waylight/internal/modules/nextcloud/usecase"
	"waylight/internal/shared/kernel"
)

const refreshInterval = 5 * time.Minute

// Module wires the Nextcloud rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Nextcloud
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
	// repo reads the settings the integrations store copies (#610).
	repo usecase.SettingsRepository
}

// New builds the Nextcloud module; cipher encrypts the app password, conns
// resolves the named connection a widget chose (#610) and signIns (may be
// nil) hands over sign-ins of the setup assistant (#654).
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, conns modules.ConnectionResolver, clock kernel.Clock, signIns kernel.SignIns) *Module {
	api := gateway.NewAPI(client)
	repo := gateway.NewSettingsRepository(db, cipher)
	uc := usecase.NewNextcloud(repo, api, connections{conns}, clock, signIns)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), repo: repo, specs: connectionSpecs(api)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "nextcloud" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"nextcloud"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/nextcloud/settings", m.handlers.Settings)
	r.Protected("PUT /api/nextcloud/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/nextcloud/status", m.handlers.Status)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "nextcloud.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

// ConnectionStatus implements modules.StatusReporter (#599).
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	_, err := m.uc.Status(ctx, "")
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

// SavedConnections implements modules.ConnectionSource (#610).
func (m *Module) SavedConnections(ctx context.Context) ([]modules.SavedConnection, error) {
	return savedConnections(ctx, m.repo)
}
