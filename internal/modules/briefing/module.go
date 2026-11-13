package briefing

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/briefing/adapter/controller"
	"trisse-home/internal/modules/briefing/adapter/gateway"
	"trisse-home/internal/modules/briefing/usecase"
	"trisse-home/internal/shared/kernel"
)

// The read ports, the push channel and their fact types are re-exported so
// the composition root can implement them without importing the use case
// package (ADR 0016).
type (
	// Sources bundles the optional read ports.
	Sources = usecase.Sources
	// Publisher delivers the daily push.
	Publisher = usecase.Publisher
	// TodoFact is an open to-do with a due date.
	TodoFact = usecase.TodoFact
	// EventFact is an upcoming appointment.
	EventFact = usecase.EventFact
	// WarningFact is an active weather warning.
	WarningFact = usecase.WarningFact
	// CollectionFact is a waste collection.
	CollectionFact = usecase.CollectionFact
	// ServiceFact is a service that is down.
	ServiceFact = usecase.ServiceFact
	// IncidentFact is a recorded failure or warning.
	IncidentFact = usecase.IncidentFact
)

// Config holds the wiring of the module.
type Config struct {
	Sources   Sources
	Publisher Publisher // nil when notifications are disabled
	// DefaultLat/Lon is the weather location until one is saved.
	DefaultLat, DefaultLon float64
}

// Module wires the briefing rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	send     usecase.SendBriefing
}

// pushCheckInterval is how often the push schedule is checked (#70).
const pushCheckInterval = time.Minute

// New builds the briefing module.
func New(db *sql.DB, clock kernel.Clock, cfg Config) *Module {
	d := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, Sources: cfg.Sources, Publisher: cfg.Publisher,
		Location: time.Local, DefaultLat: cfg.DefaultLat, DefaultLon: cfg.DefaultLon}
	send := usecase.NewSendBriefing(d)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, send: send,
		handlers: controller.NewHandlers(controller.UseCases{Build: usecase.NewBuildBriefing(d), Send: send, Settings: usecase.NewManageSettings(d)})}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "briefing" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"briefing"} }

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "briefing.push", Interval: pushCheckInterval, Run: func(ctx context.Context) error {
		_, err := m.send.Execute(ctx)
		return err
	}}}
}

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/briefing", h.Get)
	r.Protected("GET /api/briefing/settings", h.GetSettings)
	r.Protected("PUT /api/briefing/settings", h.SaveSettings)
	r.Protected("POST /api/briefing/send", h.Send)
}
