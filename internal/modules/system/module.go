package system

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/system/adapter/controller"
	"trisse-home/internal/modules/system/adapter/gateway"
	"trisse-home/internal/modules/system/domain"
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
	// Version is the running build; Client looks up new releases (only
	// when the user enabled the update check).
	Version string
	Client  *http.Client
	// InstallMethod (INSTALL_METHOD) selects the update hint.
	InstallMethod string
}

// updateCheckJobInterval: the use case itself limits lookups to once a day.
const updateCheckJobInterval = 6 * time.Hour

// Module wires the system rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Monitor
	versions usecase.Versions
	handlers *controller.Handlers
	version  *controller.VersionHandlers
}

// New builds the system module.
func New(cfg Config, db *sql.DB, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	host := gateway.NewHost(cfg.ProcPath, cfg.SysPath, gateway.ParseDiskMounts(cfg.DiskPaths))
	uc := usecase.NewMonitor(host, gateway.NewHistoryStore(db), sink, clock)
	versions := usecase.NewVersions(cfg.Version, domain.ParseInstallMethod(cfg.InstallMethod), gateway.NewGitHubReleases(cfg.Client, gateway.DefaultReleasesURL), gateway.NewUpdateSettings(db), clock)
	return &Module{uc: uc, versions: versions, handlers: controller.NewHandlers(uc), version: controller.NewVersionHandlers(versions)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "system" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"system"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/system", m.handlers.Snapshot)
	r.Protected("GET /api/system/history", m.handlers.History)
	r.Protected("GET /api/system/version", m.version.Version)
	r.Protected("PUT /api/system/update-check", m.version.SetUpdateCheck)
}

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{
		{Name: "system.sample", Interval: sampleInterval, Run: m.uc.Sample},
		// Does nothing unless the user enabled the update check.
		{Name: "system.update-check", Interval: updateCheckJobInterval, Run: m.versions.Check},
	}
}
