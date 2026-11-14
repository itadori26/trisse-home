package device

import (
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/device/adapter/controller"
	"trisse-home/internal/modules/device/adapter/gateway"
	"trisse-home/internal/modules/device/usecase"
)

// Config are the dependencies of the module.
type Config struct {
	// Appliance is true on the appliance image (INSTALL_METHOD=appliance).
	Appliance bool
	// Dir is the state directory shared with the root side.
	Dir string
	// Sysfs is /sys/class/drm.
	Sysfs fs.FS
	Kiosk usecase.KioskIssuer
}

// Module wires the device rings (#140) and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the module.
func New(c Config) *Module {
	uc := usecase.New(usecase.Deps{Appliance: c.Appliance, Detector: gateway.NewSysfs(c.Sysfs), Store: gateway.NewFiles(c.Dir), Kiosk: c.Kiosk})
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "device" }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/device", m.handlers.State)
	r.Protected("PUT /api/device/display-mode", m.handlers.ChangeMode)
}
