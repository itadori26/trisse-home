package system

import (
	"database/sql"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/system/adapter/controller"
	"trisse-home/internal/modules/system/adapter/gateway"
	"trisse-home/internal/modules/system/usecase"
	"trisse-home/internal/shared/kernel"
)

// sampleInterval balances fresh CPU numbers against wake-ups on the host.
const sampleInterval = 5 * time.Second

// Config locates the host's procfs, sysfs and the file systems to report.
type Config struct {
	ProcPath  string
	SysPath   string
	DiskPaths []string
}

// Module wires the system rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Monitor
	handlers *controller.Handlers
}

// New builds the system module.
func New(cfg Config, db *sql.DB, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	host := gateway.NewHost(cfg.ProcPath, cfg.SysPath, gateway.ParseDiskMounts(cfg.DiskPaths))
	uc := usecase.NewMonitor(host, gateway.NewHistoryStore(db), sink, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "system" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"system"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/system", m.handlers.Snapshot)
	r.Protected("GET /api/system/history", m.handlers.History)
}

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "system.sample", Interval: sampleInterval, Run: m.uc.Sample}}
}
