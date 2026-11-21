package chores

import (
	"context"

	"waylight/internal/shared/kernel"
)

// SeedDemo implements modules.DemoSeeder: household chores for the public
// demo, stored for the demo account.
func (m *Module) SeedDemo(ctx context.Context) error {
	return m.seedDemo(kernel.WithPrincipal(ctx, kernel.AdminPrincipal()))
}
