// Package usecase contains the timeline use cases.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/timeline/domain"
	"trisse-home/internal/shared/kernel"
)

// Filter selects events; empty fields match everything.
type Filter struct {
	Source   string
	Severity domain.Severity
	Before   time.Time
	Limit    int
}

// Timeline is the input port. It receives the same observations and alerts
// as the notifications module.
type Timeline interface {
	kernel.ObservationSink
	kernel.AlertSink
	List(ctx context.Context, f Filter) ([]domain.Event, error)
	RetentionDays(ctx context.Context) (int, error)
	SetRetentionDays(ctx context.Context, days int) error
	// Prune deletes events older than the retention (job).
	Prune(ctx context.Context) error
}

// Repository persists events and the retention (output port).
type Repository interface {
	Add(ctx context.Context, e domain.Event) error
	List(ctx context.Context, f Filter) ([]domain.Event, error)
	DeleteBefore(ctx context.Context, t time.Time) error
	// RetentionDays returns 0 when nothing is stored.
	RetentionDays(ctx context.Context) (int, error)
	SetRetentionDays(ctx context.Context, days int) error
}

type timeline struct {
	repo  Repository
	clock kernel.Clock
	ids   kernel.IDGenerator

	mu     sync.Mutex
	tracks map[string]domain.Track
}

// NewTimeline returns the timeline use cases.
func NewTimeline(repo Repository, clock kernel.Clock, ids kernel.IDGenerator) Timeline {
	return &timeline{repo: repo, clock: clock, ids: ids, tracks: map[string]domain.Track{}}
}

// Observe implements kernel.ObservationSink: confirmed state changes become events.
func (uc *timeline) Observe(ctx context.Context, o kernel.Observation) {
	healthy := domain.HealthyFor(o)
	key := o.Kind + ":" + o.Key
	uc.mu.Lock()
	tr, changed := uc.tracks[key].Advance(healthy)
	uc.tracks[key] = tr
	uc.mu.Unlock()
	if changed {
		_ = uc.repo.Add(ctx, domain.FromObservation(uc.ids.NewID(), uc.clock.Now(), o, healthy)) // best effort
	}
}

// Alert implements kernel.AlertSink.
func (uc *timeline) Alert(ctx context.Context, a kernel.Alert) {
	sev := domain.SeverityWarning
	if a.Urgent {
		sev = domain.SeverityCritical
	}
	msg := a.Title
	if a.Body != "" {
		msg += ": " + a.Body
	}
	_ = uc.repo.Add(ctx, domain.NewEvent(uc.ids.NewID(), uc.clock.Now(), a.Kind, sev, "", msg)) // best effort
}

// List implements Timeline (newest first).
func (uc *timeline) List(ctx context.Context, f Filter) ([]domain.Event, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	list, err := uc.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return list, nil
}

// RetentionDays implements Timeline.
func (uc *timeline) RetentionDays(ctx context.Context) (int, error) {
	d, err := uc.repo.RetentionDays(ctx)
	if err != nil {
		return 0, fmt.Errorf("load retention: %w", err)
	}
	if d == 0 {
		d = domain.DefaultRetentionDays
	}
	return d, nil
}

// SetRetentionDays implements Timeline.
func (uc *timeline) SetRetentionDays(ctx context.Context, days int) error {
	if err := domain.ValidRetention(days); err != nil {
		return err
	}
	return uc.repo.SetRetentionDays(ctx, days)
}

// Prune implements Timeline.
func (uc *timeline) Prune(ctx context.Context) error {
	days, err := uc.RetentionDays(ctx)
	if err != nil {
		return err
	}
	return uc.repo.DeleteBefore(ctx, uc.clock.Now().AddDate(0, 0, -days))
}
