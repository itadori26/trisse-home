package timeline

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/timeline/adapter/controller"
	"trisse-home/internal/modules/timeline/adapter/gateway"
	"trisse-home/internal/modules/timeline/domain"
	"trisse-home/internal/modules/timeline/usecase"
	"trisse-home/internal/shared/kernel"
)

// pruneInterval: old events are removed once a day.
const pruneInterval = 24 * time.Hour

// Module wires the timeline rings, implements modules.Module and receives
// observations and alerts like the notifications module.
type Module struct {
	modules.Base
	uc       usecase.Timeline
	handlers *controller.Handlers
}

// New builds the timeline module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewTimeline(gateway.NewRepository(db), clock, ids)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Observe implements kernel.ObservationSink.
func (m *Module) Observe(ctx context.Context, o kernel.Observation) { m.uc.Observe(ctx, o) }

// Alert implements kernel.AlertSink.
func (m *Module) Alert(ctx context.Context, a kernel.Alert) { m.uc.Alert(ctx, a) }

// Name implements modules.Module.
func (m *Module) Name() string { return "timeline" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"timeline"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/timeline", m.handlers.List)
	r.Protected("GET /api/timeline/settings", m.handlers.Settings)
	r.Protected("PUT /api/timeline/settings", m.handlers.SaveSettings)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "timeline.prune", Interval: pruneInterval, Run: m.uc.Prune}}
}

// Recent returns up to limit events, newest first (read access for the
// briefing, ADR 0016).
func (m *Module) Recent(ctx context.Context, limit int) ([]domain.Event, error) {
	return m.uc.List(ctx, usecase.Filter{Limit: limit})
}
