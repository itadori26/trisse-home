package backup

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/backup/adapter/controller"
	"waylight/internal/modules/backup/adapter/gateway"
	"waylight/internal/modules/backup/domain"
	"waylight/internal/modules/backup/usecase"
	"waylight/internal/shared/kernel"
)

// checkInterval is how often the job checks whether a backup is due.
const checkInterval = time.Hour

// uploadTimeout bounds one off-site upload.
const uploadTimeout = 30 * time.Minute

// MaxRestoreBytes limits an uploaded or downloaded backup for a restore.
const MaxRestoreBytes = 2 << 30

// Config holds the backup settings.
type Config struct {
	Dir       string
	Interval  time.Duration
	Retention int
	// Off-site backups (#46): Cipher encrypts the WebDAV password, Client
	// uploads (only once a target is configured), Alerts get failures.
	Cipher gateway.Cipher
	Client *http.Client
	Alerts kernel.AlertSink
	Log    *slog.Logger
	// Restore (#46): Swapper stages and schedules the database swap,
	// Reauth re-checks password and 2FA, Restart restarts the server.
	Swapper usecase.DatabaseSwapper
	Reauth  usecase.Reauthenticator
	Restart func()
}

// Module wires the backup rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Backups
	clock    kernel.Clock
	handlers *controller.Handlers
	remote   *controller.RemoteHandlers
	restore  *controller.RestoreHandlers
}

// New builds the backup module.
func New(cfg Config, db *sql.DB, clock kernel.Clock) (*Module, error) {
	policy, err := domain.NewPolicy(cfg.Interval, cfg.Retention)
	if err != nil {
		return nil, fmt.Errorf("backup policy: %w", err)
	}
	store, err := gateway.NewDirStore(cfg.Dir)
	if err != nil {
		return nil, err
	}
	offsite := usecase.NewOffsite(store, gateway.NewWebDAV(cfg.Client), gateway.NewRemoteSettingsRepository(db, cfg.Cipher), cfg.Alerts)
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	tracker := &usecase.Tracker{}
	// Uploads run in the background so a manual backup returns at once.
	upload := func(ctx context.Context, f usecase.FileView) {
		go func() {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), uploadTimeout)
			defer cancel()
			err := offsite.Upload(ctx, f.Name)
			tracker.OffsiteResult(clock.Now(), err)
			if err != nil {
				log.Warn("off-site backup failed", "file", f.Name, "error", err)
			}
		}()
	}
	uc := usecase.NewBackups(policy, gateway.NewSnapshotter(db), store, clock,
		usecase.WithAfterBackup(upload), usecase.WithTracker(tracker))
	m := &Module{uc: uc, clock: clock, handlers: controller.NewHandlers(uc), remote: controller.NewRemoteHandlers(offsite)}
	if cfg.Swapper != nil && cfg.Reauth != nil && cfg.Restart != nil {
		restorer := usecase.NewRestorer(usecase.RestoreDeps{Store: store, Remote: gateway.NewWebDAV(cfg.Client),
			RemoteCfg: gateway.NewRemoteSettingsRepository(db, cfg.Cipher), Inspector: gateway.NewInspector(db),
			Swapper: cfg.Swapper, Reauth: cfg.Reauth, Alerts: cfg.Alerts, Restart: cfg.Restart})
		m.restore = controller.NewRestoreHandlers(restorer, MaxRestoreBytes)
	}
	return m, nil
}

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// Name implements modules.Module.
func (m *Module) Name() string { return "backup" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/backups", m.handlers.List)
	r.Protected("GET /api/backups/status", m.handlers.Status)
	r.Protected("POST /api/backups", m.handlers.Create)
	r.Protected("GET /api/backups/remote", m.remote.List)
	r.Protected("GET /api/backups/remote/settings", m.remote.Settings)
	r.Protected("PUT /api/backups/remote/settings", m.remote.SaveSettings)
	r.Protected("DELETE /api/backups/remote/settings", m.remote.Disable)
	r.Protected("POST /api/backups/remote/test", m.remote.Test)
	if m.restore != nil {
		r.Protected("POST /api/backups/restore", m.restore.Restore)
	}
	r.Protected("GET /api/backups/{name}", m.handlers.Download)
}

// WidgetKinds implements modules.Module: the traffic light of Waylight's
// own backups (#677).
func (m *Module) WidgetKinds() []string { return []string{"backup-health"} }

// CompactState implements modules.CompactProvider (#677): the traffic light
// of the backups in one line.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	s, err := m.uc.Status(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	return usecase.CompactState(s, m.clock.Now(), time.Local), nil
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "backup.scheduled", Interval: checkInterval, Run: m.uc.RunScheduled}}
}
