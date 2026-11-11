package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"trisse-home/internal/platform/scheduler"
	"trisse-home/internal/shared/sqlutil"
)

// prepareDemo creates the demo account (which seeds the sample data), takes
// a snapshot of all exportable data and returns the job that restores it
// every DEMO_RESET_INTERVAL.
func prepareDemo(ctx context.Context, a *app) (scheduler.Task, error) {
	if err := a.auth.PrepareDemo(ctx); err != nil {
		return scheduler.Task{}, err
	}
	snapshot := map[string]map[string]sqlutil.Rows{}
	for name, p := range a.registry.Portables() {
		data, err := p.ExportData(ctx)
		if err != nil {
			return scheduler.Task{}, fmt.Errorf("snapshot %s: %w", name, err)
		}
		snapshot[name] = data
	}
	a.log.Warn("DEMO_MODE: public demo without login; data is reset regularly", "reset_interval", a.cfg.DemoResetInterval)
	// The scheduler runs every task once at start; the data is fresh then,
	// and resetting would race with the first feed and weather refresh.
	var started atomic.Bool
	return scheduler.Task{Name: "demo.reset", Interval: a.cfg.DemoResetInterval, Run: func(ctx context.Context) error {
		if started.CompareAndSwap(false, true) {
			return nil
		}
		for name, p := range a.registry.Portables() {
			if err := p.ImportData(ctx, snapshot[name]); err != nil {
				return fmt.Errorf("reset %s: %w", name, err)
			}
		}
		a.log.InfoContext(ctx, "demo data reset")
		return nil
	}}, nil
}
