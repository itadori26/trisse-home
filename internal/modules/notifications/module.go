package notifications

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/notifications/adapter/controller"
	"trisse-home/internal/modules/notifications/adapter/gateway"
	"trisse-home/internal/modules/notifications/domain"
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
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewNotifier(usecase.Deps{
		Repo: gateway.NewRepository(db, cipher), Sender: gateway.NewNtfySender(client),
		Channels: gateway.NewChannelRepository(db, cipher), Maintenance: gateway.NewMaintenanceRepository(db),
		Pending: gateway.NewPendingRepository(db), Clock: clock, IDs: ids, Location: time.Local,
		Publishers: map[domain.ChannelKind]usecase.Publisher{
			domain.ChannelGotify:  gateway.NewGotifyPublisher(client),
			domain.ChannelWebhook: gateway.NewWebhookPublisher(client),
			domain.ChannelEmail:   gateway.NewEmailPublisher(clock.Now),
		},
	})
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
	r.Protected("GET /api/notifications/channels", m.handlers.Channels)
	r.Protected("POST /api/notifications/channels", m.handlers.SaveChannel)
	r.Protected("PUT /api/notifications/channels/{id}", m.handlers.SaveChannel)
	r.Protected("DELETE /api/notifications/channels/{id}", m.handlers.DeleteChannel)
	r.Protected("POST /api/notifications/channels/{id}/test", m.handlers.TestChannel)
	r.Protected("GET /api/notifications/maintenance", m.handlers.Maintenance)
	r.Protected("POST /api/notifications/maintenance", m.handlers.StartMaintenance)
	r.Protected("DELETE /api/notifications/maintenance/{id}", m.handlers.EndMaintenance)
}

// digestInterval: how often deferred notifications are checked after quiet hours.
const digestInterval = time.Minute

// Jobs implements modules.Module: sends the quiet-hours digest.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "notifications.digest", Interval: digestInterval, Run: m.uc.FlushDigest}}
}

// UnderMaintenance reports whether an object is muted by a maintenance
// window (used by the services module to exclude it from uptime, #81).
func (m *Module) UnderMaintenance(ctx context.Context, kind, key, label string, at time.Time) bool {
	return m.uc.UnderMaintenance(ctx, kind, key, label, at)
}
