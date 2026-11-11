package backup

import (
	"database/sql"
	"fmt"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/backup/adapter/controller"
	"trisse-home/internal/modules/backup/adapter/gateway"
	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/modules/backup/usecase"
	"trisse-home/internal/shared/kernel"
)

// checkInterval is how often the job checks whether a backup is due.
const checkInterval = time.Hour

// Config holds the backup settings.
type Config struct {
	Dir       string
	Interval  time.Duration
	Retention int
}

// Module wires the backup rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Backups
	handlers *controller.Handlers
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
	uc := usecase.NewBackups(policy, gateway.NewSnapshotter(db), store, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}, nil
}

// Name implements modules.Module.
func (m *Module) Name() string { return "backup" }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/backups", m.handlers.List)
	r.Protected("POST /api/backups", m.handlers.Create)
	r.Protected("GET /api/backups/{name}", m.handlers.Download)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "backup.scheduled", Interval: checkInterval, Run: m.uc.RunScheduled}}
}
