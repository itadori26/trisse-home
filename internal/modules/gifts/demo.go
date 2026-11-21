package gifts

import (
	"context"

	"waylight/internal/shared/kernel"
)

// SeedDemo implements modules.DemoSeeder: ideas for the people of the demo
// birthdays, stored for the demo account.
func (m *Module) SeedDemo(ctx context.Context) error {
	return m.seedDemo(kernel.WithPrincipal(ctx, kernel.AdminPrincipal()))
}
