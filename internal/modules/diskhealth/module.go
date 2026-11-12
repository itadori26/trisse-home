package diskhealth

import (
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/diskhealth/adapter/controller"
	"trisse-home/internal/modules/diskhealth/adapter/gateway"
	"trisse-home/internal/modules/diskhealth/usecase"
	"trisse-home/internal/shared/kernel"
)

// evaluateInterval: how often disk states are reported to notifications.
const evaluateInterval = 5 * time.Minute

// Config configures the module.
type Config struct {
	// Dir is the export directory of scripts/smart-export.sh.
	Dir string
	// StaleAfter marks older exports as unknown (0 = default).
	StaleAfter time.Duration
}

// Module wires the disk health rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.DiskHealth
	handlers *controller.Handlers
}

// New builds the disk health module; alerter receives state observations.
func New(cfg Config, alerter kernel.ObservationSink, clock kernel.Clock) *Module {
	uc := usecase.NewDiskHealth(gateway.NewFileSource(cfg.Dir), alerter, clock, cfg.StaleAfter)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "diskhealth" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"diskhealth"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/diskhealth", m.handlers.List)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "diskhealth.evaluate", Interval: evaluateInterval, Run: m.uc.Evaluate}}
}
