package services

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/services/adapter/controller"
	"trisse-home/internal/modules/services/adapter/gateway"
	"trisse-home/internal/modules/services/usecase"
	"trisse-home/internal/shared/kernel"
)

const checkInterval = time.Minute

// Deps are the infrastructure services the module needs.
type Deps struct {
	DB             *sql.DB
	Client         *http.Client
	InsecureClient *http.Client
	Clock          kernel.Clock
	IDs            kernel.IDGenerator
}

// Module wires the services rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Services
	seed     usecase.SeedSampleData
	handlers *controller.Handlers
}

// New builds the services module.
func New(d Deps) *Module {
	deps := usecase.Deps{
		Repo: gateway.NewRepository(d.DB), Status: gateway.NewStatusStore(),
		Checker: gateway.NewHTTPChecker(d.Client, d.InsecureClient), History: gateway.NewHistoryStore(d.DB), Clock: d.Clock, IDs: d.IDs,
	}
	uc := usecase.NewServices(deps)
	return &Module{uc: uc, seed: usecase.NewSeedSampleData(deps), handlers: controller.NewHandlers(uc)}
}

// SeedSampleData implements modules.Seeder. The example tiles are checked in
// the background so they show their status right away without delaying setup.
func (m *Module) SeedSampleData(ctx context.Context) error {
	created, err := m.seed.Execute(ctx)
	if err != nil || !created {
		return err
	}
	go func() {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), checkInterval)
		defer cancel()
		_ = m.uc.CheckAll(bg) // failures show up as offline tiles and are retried by the job
	}()
	return nil
}

// Name implements modules.Module.
func (m *Module) Name() string { return "services" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"services"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/services", h.List)
	r.Protected("POST /api/services", h.Create)
	r.Protected("POST /api/services/check", h.Check)
	r.Protected("GET /api/services/history", h.History)
	r.Protected("PUT /api/services/order", h.Reorder)
	r.Protected("PUT /api/services/{id}", h.Update)
	r.Protected("DELETE /api/services/{id}", h.Delete)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "services.health-check", Interval: checkInterval, Run: func(ctx context.Context) error {
		return m.uc.CheckAll(ctx)
	}}}
}
