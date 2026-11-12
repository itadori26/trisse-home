// Package usecase contains the disk health use cases.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"trisse-home/internal/modules/diskhealth/domain"
	"trisse-home/internal/shared/kernel"
)

// DiskHealth is the input port.
type DiskHealth interface {
	// Disks returns the evaluated disks (GetDiskHealth).
	Disks(ctx context.Context) ([]domain.Disk, error)
	// Evaluate reports every disk's state to the alerter (job).
	Evaluate(ctx context.Context) error
}

// SmartReportSource reads the exported smartctl reports (output port).
type SmartReportSource interface {
	Reports(ctx context.Context) ([]domain.Report, error)
}

type diskHealth struct {
	source     SmartReportSource
	alerter    kernel.ObservationSink // HealthAlerter
	clock      kernel.Clock
	staleAfter time.Duration
}

// NewDiskHealth returns the use cases; staleAfter <= 0 uses the default.
func NewDiskHealth(source SmartReportSource, alerter kernel.ObservationSink, clock kernel.Clock, staleAfter time.Duration) DiskHealth {
	if staleAfter <= 0 {
		staleAfter = domain.DefaultStaleAfter
	}
	if alerter == nil {
		alerter = kernel.DiscardObservations{}
	}
	return &diskHealth{source: source, alerter: alerter, clock: clock, staleAfter: staleAfter}
}

// Disks implements DiskHealth (sorted by device).
func (uc *diskHealth) Disks(ctx context.Context) ([]domain.Disk, error) {
	reports, err := uc.source.Reports(ctx)
	if err != nil {
		return nil, err
	}
	now := uc.clock.Now()
	out := make([]domain.Disk, 0, len(reports))
	for _, r := range reports {
		out = append(out, domain.Evaluate(r, now, uc.staleAfter))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Device < out[j].Device })
	return out, nil
}

// Evaluate implements DiskHealth.
func (uc *diskHealth) Evaluate(ctx context.Context) error {
	disks, err := uc.Disks(ctx)
	if err != nil {
		if errors.Is(err, kernel.ErrNotFound) {
			return nil // not set up: nothing to report
		}
		return err
	}
	for _, d := range disks {
		if o, ok := d.Observation(); ok {
			uc.alerter.Observe(ctx, o)
		}
	}
	return nil
}
