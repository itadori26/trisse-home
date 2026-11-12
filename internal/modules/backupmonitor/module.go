package backupmonitor

import (
	"database/sql"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/backupmonitor/adapter/controller"
	"trisse-home/internal/modules/backupmonitor/adapter/gateway"
	"trisse-home/internal/modules/backupmonitor/usecase"
	"trisse-home/internal/shared/kernel"
)

// reportInterval is how often the source states are reported for notifications.
const reportInterval = 5 * time.Minute

// Deps are the services the module needs.
type Deps struct {
	DB           *sql.DB
	Tokens       usecase.TokenGenerator
	Clock        kernel.Clock
	IDs          kernel.IDGenerator
	Observations kernel.ObservationSink
}

// Module wires the backup monitor rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Monitor
	handlers *controller.Handlers
}

// New builds the backup monitor module.
func New(d Deps) *Module {
	uc := usecase.NewMonitor(usecase.Deps{Repo: gateway.NewRepository(d.DB), Tokens: d.Tokens, Clock: d.Clock, IDs: d.IDs, Observations: d.Observations})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "backupmonitor" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"backup-status"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	// /api/checks is the general name (#100); /api/backup-status stays as an
	// alias so existing backup scripts and ping URLs keep working.
	for _, base := range []string{"/api/backup-status", "/api/checks"} {
		r.Protected("GET "+base+"/sources", h.List)
		r.Protected("POST "+base+"/sources", h.Create)
		r.Protected("PUT "+base+"/sources/{id}", h.Update)
		r.Protected("DELETE "+base+"/sources/{id}", h.Delete)
		r.Protected("POST "+base+"/sources/{id}/token", h.RotateToken)
		for _, method := range []string{"GET", "POST"} {
			r.Public(method+" "+base+"/ping/{token}", h.Ping)
			r.Public(method+" "+base+"/ping/{token}/{event}", h.Ping)
		}
	}
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "backupmonitor.report", Interval: reportInterval, Run: m.uc.Report}}
}
