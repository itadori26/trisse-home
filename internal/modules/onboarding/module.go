package onboarding

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/onboarding/adapter/controller"
	"trisse-home/internal/modules/onboarding/adapter/gateway"
	"trisse-home/internal/modules/onboarding/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the onboarding rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Onboarding
	handlers *controller.Handlers
}

// New builds the onboarding module.
func New(db *sql.DB, clock kernel.Clock) *Module {
	uc := usecase.NewOnboarding(gateway.NewRepository(db), clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "onboarding" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/onboarding", m.handlers.Status)
	r.Protected("PUT /api/onboarding", m.handlers.Update)
}

// RestartTour makes the tour show again, e.g. after a new account was set up.
func (m *Module) RestartTour(ctx context.Context) error {
	_, err := m.uc.Restart(ctx)
	return err
}
