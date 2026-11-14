package main

import (
	"context"
	"os"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/modules/device"
)

// The device module registers itself (#193); it only works on the
// appliance image (#140).
func init() {
	registerModule("device", func(d moduleDeps) modules.Module {
		return device.New(device.Config{
			Appliance: d.cfg.InstallMethod == "appliance" && !d.cfg.DemoMode,
			Dir:       d.cfg.DeviceDir,
			Sysfs:     os.DirFS("/sys/class/drm"),
			Kiosk:     deviceKiosk{auth: d.auth, registry: d.registry},
		})
	})
}

// tabLister is implemented by the dashboard module.
type tabLister interface {
	TabIDs(ctx context.Context) ([]string, error)
}

// deviceKiosk implements the device module's KioskIssuer with the auth
// module: the appliance's own display shows all tabs read-only, signed in
// with a link that is accepted from the device itself only.
type deviceKiosk struct {
	auth     *auth.Module
	registry *modules.Registry
}

// Link implements device's usecase.KioskIssuer.
func (k deviceKiosk) Link(ctx context.Context) (string, error) {
	var tabs []string
	for _, m := range k.registry.All() {
		if l, ok := m.(tabLister); ok {
			ids, err := l.TabIDs(ctx)
			if err != nil {
				return "", err
			}
			tabs = ids
		}
	}
	return k.auth.DeviceKioskLink(ctx, tabs)
}

// Remove implements device's usecase.KioskIssuer.
func (k deviceKiosk) Remove(ctx context.Context) error { return k.auth.RemoveDeviceKiosk(ctx) }
