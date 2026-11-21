package arr

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/arr/adapter/controller"
	"waylight/internal/modules/arr/adapter/gateway"
	"waylight/internal/modules/arr/domain"
	"waylight/internal/modules/arr/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the *arr rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Arr
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
	// repo reads the settings the integrations store copies (#610).
	repo usecase.InstanceRepository
}

// New builds the *arr module (#86); cipher encrypts the API keys, conns
// resolves the named connections a widget chose (#610).
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, conns modules.ConnectionResolver, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	providers := map[domain.App]usecase.Provider{
		domain.AppSonarr: gateway.NewSonarr(client), domain.AppRadarr: gateway.NewRadarr(client), domain.AppProwlarr: gateway.NewProwlarr(client),
	}
	repo := gateway.NewInstanceRepository(db, cipher)
	uc := usecase.NewArr(usecase.Deps{Repo: repo, Providers: providers, Conns: connections{conns}, Clock: clock, IDs: ids})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), repo: repo, specs: connectionSpecs(providers)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "arr" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"arr"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/arr/instances", m.handlers.Instances)
	r.Protected("POST /api/arr/instances", m.handlers.Add)
	r.Protected("PUT /api/arr/instances/{id}", m.handlers.Update)
	r.Protected("DELETE /api/arr/instances/{id}", m.handlers.Delete)
	r.Protected("GET /api/arr/overview", m.handlers.Overview)
}

// ConnectionStatus implements modules.StatusReporter (#592): the cached
// overview names the instances that did not answer.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	o, err := m.uc.Overview(ctx, nil)
	if err == nil && len(o.Errors) > 0 {
		err = errInstanceDown
	}
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// errInstanceDown is the reason when one of several instances fails; the
// widget names which one.
var errInstanceDown = kernel.NewPublicError(kernel.ErrUnavailable, "A Sonarr, Radarr or Prowlarr instance does not answer.")

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

// SavedConnections implements modules.ConnectionSource (#610).
func (m *Module) SavedConnections(ctx context.Context) ([]modules.SavedConnection, error) {
	return savedConnections(ctx, m.repo)
}
