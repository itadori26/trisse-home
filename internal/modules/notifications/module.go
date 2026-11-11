package notifications

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/notifications/adapter/controller"
	"trisse-home/internal/modules/notifications/adapter/gateway"
	"trisse-home/internal/modules/notifications/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the notification rings, implements modules.Module and is the
// kernel.ObservationSink other modules report to.
type Module struct {
	modules.Base
	uc       usecase.Notifier
	handlers *controller.Handlers
}

// New builds the notifications module; cipher encrypts the ntfy token.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewNotifier(gateway.NewRepository(db, cipher), gateway.NewNtfySender(client), clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Alert implements kernel.AlertSink.
func (m *Module) Alert(ctx context.Context, a kernel.Alert) { m.uc.Alert(ctx, a) }

// Observe implements kernel.ObservationSink.
func (m *Module) Observe(ctx context.Context, o kernel.Observation) { m.uc.Observe(ctx, o) }

// Name implements modules.Module.
func (m *Module) Name() string { return "notifications" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/notifications/settings", m.handlers.Settings)
	r.Protected("PUT /api/notifications/settings", m.handlers.SaveSettings)
	r.Protected("POST /api/notifications/test", m.handlers.SendTest)
}
