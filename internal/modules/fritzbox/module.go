package fritzbox

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules/fritzbox/domain"

	"waylight/internal/modules"
	"waylight/internal/modules/fritzbox/adapter/controller"
	"waylight/internal/modules/fritzbox/adapter/gateway"
	"waylight/internal/modules/fritzbox/usecase"
	"waylight/internal/shared/kernel"
)

// refreshInterval keeps the throughput reasonably current. The call list is
// not part of the background refresh; a calls widget asks for it (#168).
const refreshInterval = 15 * time.Second

// Module wires the FRITZ!Box rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.FritzBox
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
	// repo reads the settings the integrations store copies (#610).
	repo usecase.SettingsRepository
}

// New builds the FRITZ!Box module; cipher encrypts the password, conns
// resolves the named connection a widget chose (#610) and loc is the time
// zone of the call list.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, conns modules.ConnectionResolver, clock kernel.Clock, loc *time.Location) *Module {
	box := gateway.NewBox(client, loc)
	repo := gateway.NewSettingsRepository(db, cipher)
	uc := usecase.NewFritzBox(repo, box, connections{conns}, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), repo: repo, specs: connectionSpecs(box)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "fritzbox" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module: the internet connection and,
// separately, the call list (#168).
func (m *Module) WidgetKinds() []string { return []string{"fritzbox", "fritzbox-calls"} }

// GuestGatedRoutes implements modules.GuestGated: call data is personal, so
// a guest reads it only when a calls widget is on one of its tabs (#168).
func (m *Module) GuestGatedRoutes() map[string]string {
	return map[string]string{"/api/fritzbox/calls": "fritzbox-calls"}
}

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/fritzbox/settings", m.handlers.Settings)
	r.Protected("PUT /api/fritzbox/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/fritzbox/status", m.handlers.Status)
	r.Protected("GET /api/fritzbox/calls", m.handlers.Calls)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "fritzbox.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

// ConnectionStatus implements modules.StatusReporter (#599).
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	_, err := m.uc.Status(ctx, "")
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// Discovery implements modules.Discoverable (#627): the box is the
// gateway or fritz.box, and its TR-064 description is public.
func (m *Module) Discovery() modules.DiscoverySpec {
	return modules.DiscoverySpec{
		Service: "FRITZ!Box", Hostnames: []string{"fritz.box"},
		Ports: []int{49000}, ProbePath: "/tr64desc.xml", ProbeContains: "AVM",
	}
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

// SavedConnections implements modules.ConnectionSource (#610).
func (m *Module) SavedConnections(ctx context.Context) ([]modules.SavedConnection, error) {
	return savedConnections(ctx, m.repo)
}
