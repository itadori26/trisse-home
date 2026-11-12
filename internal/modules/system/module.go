package system

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
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
	// Version is the running build; Client looks up new releases (only
	// when the user enabled the update check).
	Version string
	Client  *http.Client
}

// forecastInterval: how often today's disk aggregate is updated.
const forecastInterval = 15 * time.Minute

// updateCheckJobInterval: the use case itself limits lookups to once a day.
const updateCheckJobInterval = 6 * time.Hour

// Module wires the system rings and implements modules.Module.
type Module struct {
	modules.Base
	uc        usecase.Monitor
	forecasts usecase.DiskForecasts
	versions  usecase.Versions
	handlers  *controller.Handlers
	version   *controller.VersionHandlers
}

// New builds the system module.
func New(cfg Config, db *sql.DB, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	host := gateway.NewHost(cfg.ProcPath, cfg.SysPath, gateway.ParseDiskMounts(cfg.DiskPaths))
	uc := usecase.NewMonitor(host, gateway.NewHistoryStore(db), sink, clock)
	versions := usecase.NewVersions(cfg.Version, gateway.NewGitHubReleases(cfg.Client, gateway.DefaultReleasesURL), gateway.NewUpdateSettings(db), clock)
	forecasts := usecase.NewDiskForecasts(uc, gateway.NewAggregateStore(db), sink, clock, time.Local)
	return &Module{uc: uc, forecasts: forecasts, versions: versions, handlers: controller.NewHandlers(uc, forecasts), version: controller.NewVersionHandlers(versions)}
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
	r.Protected("GET /api/system/forecast-settings", m.handlers.ForecastSettings)
	r.Protected("PUT /api/system/forecast-settings", m.handlers.SaveForecastSettings)
}

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{
		{Name: "system.sample", Interval: sampleInterval, Run: m.uc.Sample},
		// Daily usage aggregates and the disk forecast (#108).
		{Name: "system.forecast", Interval: forecastInterval, Run: m.forecasts.Aggregate},
		// Does nothing unless the user enabled the update check.
		{Name: "system.update-check", Interval: updateCheckJobInterval, Run: m.versions.Check},
	}
}

// Metrics implements kernel.MetricSource (#109): host metrics.
func (m *Module) Metrics(ctx context.Context) []kernel.MetricSample {
	s, err := m.uc.Current(ctx)
	if err != nil {
		return nil
	}
	out := []kernel.MetricSample{
		kernel.Gauge("trisse_host_cpu_ratio", "CPU usage of the host (0-1).", s.CPUPercent/100),
		kernel.Gauge("trisse_host_load1", "1-minute load average.", s.Load1),
		kernel.Gauge("trisse_host_memory_used_bytes", "Used memory of the host.", float64(s.Memory.Used)),
		kernel.Gauge("trisse_host_memory_total_bytes", "Total memory of the host.", float64(s.Memory.Total)),
		kernel.Gauge("trisse_host_uptime_seconds", "Uptime of the host.", s.Uptime.Seconds()),
	}
	if s.Temperature > 0 {
		out = append(out, kernel.Gauge("trisse_host_temperature_celsius", "Highest temperature sensor reading.", s.Temperature))
	}
	for _, d := range s.Disks {
		l := kernel.Label{Name: "disk", Value: d.Label}
		out = append(out,
			kernel.Gauge("trisse_host_disk_used_bytes", "Used bytes per configured file system.", float64(d.Usage.Used), l),
			kernel.Gauge("trisse_host_disk_total_bytes", "Size per configured file system.", float64(d.Usage.Total), l))
	}
	return out
}
