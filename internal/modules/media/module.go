package media

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/media/adapter/controller"
	"waylight/internal/modules/media/adapter/gateway"
	"waylight/internal/modules/media/domain"
	"waylight/internal/modules/media/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the media server rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Media
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
	// repo reads the settings the integrations store copies (#610).
	repo usecase.SettingsRepository
}

// New builds the media module (#61); cipher encrypts the stored token,
// conns resolves the named connection a widget chose (#610).
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, conns modules.ConnectionResolver, clock kernel.Clock) *Module {
	providers := map[domain.Provider]usecase.MediaServerProvider{
		domain.ProviderJellyfin: gateway.NewJellyfin(client),
		domain.ProviderPlex:     gateway.NewPlex(client),
	}
	repo := gateway.NewSettingsRepository(db, cipher)
	uc := usecase.NewMedia(repo, providers, connections{conns}, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), repo: repo, specs: connectionSpecs(providers)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "media" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"media"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/media/settings", m.handlers.Settings)
	r.Protected("PUT /api/media/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/media/overview", m.handlers.Overview)
	r.Protected("GET /api/media/posters/{id}", m.handlers.Poster)
}

// ConnectionStatus implements modules.StatusReporter (#592): the overview
// the widget shows tells whether the media server answers; it is cached.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	_, err := m.uc.Overview(ctx, "")
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// Discovery implements modules.Discoverable (#627). Only Jellyfin is
// found: its public system info needs no key; Plex is entered by hand.
func (m *Module) Discovery() modules.DiscoverySpec {
	return modules.DiscoverySpec{
		Service: "Jellyfin", Images: []string{"jellyfin"}, Hostnames: []string{"jellyfin", "jellyfin.local"},
		Ports: []int{8096}, ProbePath: "/System/Info/Public", ProbeContains: `"ProductName":"Jellyfin Server"`,
	}
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

// SavedConnections implements modules.ConnectionSource (#610).
func (m *Module) SavedConnections(ctx context.Context) ([]modules.SavedConnection, error) {
	return savedConnections(ctx, m.repo)
}
