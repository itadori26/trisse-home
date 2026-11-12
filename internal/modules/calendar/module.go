package calendar

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/calendar/adapter/controller"
	"trisse-home/internal/modules/calendar/adapter/gateway"
	"trisse-home/internal/modules/calendar/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = 15 * time.Minute

// Module wires the calendar rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Service
	handlers *controller.Handlers
	subs     *controller.SubscriptionHandlers
}

// New builds the calendar module. Floating event times are interpreted in loc.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location) *Module {
	caldav := gateway.NewCalDAV(client, loc)
	uc := usecase.NewCalendar(usecase.Deps{
		Repo: gateway.NewSettingsRepository(db, cipher), Source: caldav, Writer: caldav,
		Subscriptions: gateway.NewSubscriptionRepository(db, cipher), Feeds: gateway.NewFeed(client, loc, clock.Now),
		IDs: ids, Clock: clock,
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), subs: controller.NewSubscriptionHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "calendar" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"calendar"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/calendar/settings", m.handlers.Settings)
	r.Protected("PUT /api/calendar/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/calendar/events", m.handlers.Events)
	r.Protected("POST /api/calendar/events", m.handlers.CreateEvent)
	r.Protected("GET /api/calendar/subscriptions", m.subs.List)
	r.Protected("POST /api/calendar/subscriptions", m.subs.Add)
	r.Protected("PUT /api/calendar/subscriptions/{id}", m.subs.Update)
	r.Protected("DELETE /api/calendar/subscriptions/{id}", m.subs.Delete)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "calendar.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
