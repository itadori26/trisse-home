package fritzbox

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/fritzbox/adapter/controller"
	"trisse-home/internal/modules/fritzbox/adapter/gateway"
	"trisse-home/internal/modules/fritzbox/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval keeps the throughput reasonably current. The call list is
// not part of the background refresh; a calls widget asks for it (#168).
const refreshInterval = 15 * time.Second

// Module wires the FRITZ!Box rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.FritzBox
	handlers *controller.Handlers
}

// New builds the FRITZ!Box module; cipher encrypts the password and loc is
// the time zone of the call list.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, loc *time.Location) *Module {
	uc := usecase.NewFritzBox(gateway.NewSettingsRepository(db, cipher), gateway.NewBox(client, loc), clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "fritzbox" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module: the internet connection and,
// separately, the call list (#168).
func (m *Module) WidgetKinds() []string { return []string{"fritzbox", "fritzbox-calls"} }

// GuestGatedRoutes implements modules.GuestGated: call data is personal, so
// a guest reads it only when a calls widget is on one of its tabs (#168).
func (m *Module) GuestGatedRoutes() map[string]string {
	return map[string]string{"/api/fritzbox/calls": "fritzbox-calls"}
}

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/fritzbox/settings", m.handlers.Settings)
	r.Protected("PUT /api/fritzbox/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/fritzbox/status", m.handlers.Status)
	r.Protected("GET /api/fritzbox/calls", m.handlers.Calls)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "fritzbox.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
