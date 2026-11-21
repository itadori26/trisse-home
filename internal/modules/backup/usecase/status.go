package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/shared/kernel"
)

// Tracker remembers the outcome of the latest backup attempt and off-site
// upload for the traffic light (#677). It lives in memory on purpose: the
// scheduled job runs right after a start, and a failed attempt leaves no
// newer file, so the job tries again and sets the outcome anew.
type Tracker struct {
	mu sync.Mutex
	o  domain.Outcome
}

// BackupResult records a backup attempt; a success clears an earlier failure.
func (t *Tracker) BackupResult(at time.Time, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err == nil {
		t.o.FailedAt, t.o.Error = time.Time{}, ""
		return
	}
	t.o.FailedAt, t.o.Error = at, err.Error()
}

// OffsiteResult records an off-site upload; a success clears an earlier failure.
func (t *Tracker) OffsiteResult(at time.Time, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err == nil {
		t.o.OffsiteFailedAt, t.o.OffsiteError = time.Time{}, ""
		return
	}
	t.o.OffsiteFailedAt, t.o.OffsiteError = at, err.Error()
}

func (t *Tracker) outcome() domain.Outcome {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.o
}

// StatusView is the traffic light at the use case boundary; zero times
// mean "never" or "not since the start".
type StatusView struct {
	Health       domain.Health
	LastBackupAt time.Time
	Count        int
	Interval     time.Duration
	StaleAfter   time.Duration
	domain.Outcome
}

// Status implements Backups.
func (uc *backups) Status(ctx context.Context) (StatusView, error) {
	files, err := uc.store.List(ctx)
	if err != nil {
		return StatusView{}, fmt.Errorf("list backups: %w", err)
	}
	var latest time.Time
	for _, f := range files {
		if f.CreatedAt.After(latest) {
			latest = f.CreatedAt
		}
	}
	o := uc.tracker.outcome()
	return StatusView{
		Health:       uc.policy.Health(latest, o, uc.clock.Now()),
		LastBackupAt: latest,
		Count:        len(files),
		Interval:     uc.policy.Interval,
		StaleAfter:   uc.policy.StaleAfter(),
		Outcome:      o,
	}, nil
}

// compactTimeLayout is numeric so the compact text reads the same in every
// language.
const compactTimeLayout = "2006-01-02 15:04"

// CompactState sums the status up for the compact widget view (#677): red
// for a failed attempt, amber for an overdue backup, none at all or a
// failed off-site copy, green otherwise. Times are shown in loc.
func CompactState(s StatusView, now time.Time, loc *time.Location) kernel.CompactState {
	state := kernel.CompactState{UpdatedAt: now}
	last := func() kernel.Text {
		return kernel.NewTextf("Last backup {time}", "time", s.LastBackupAt.In(loc).Format(compactTimeLayout))
	}
	switch {
	case s.Health == domain.HealthFailed:
		state.Level = kernel.CompactProblem
		state.Title = kernel.NewText("The last backup failed")
		state.Detail = kernel.NewTextf("Failed {time}", "time", s.FailedAt.In(loc).Format(compactTimeLayout))
	case s.LastBackupAt.IsZero():
		state.Level = kernel.CompactAttention
		state.Title = kernel.NewText("No backup yet")
	case s.Health == domain.HealthStale && !s.OffsiteFailedAt.IsZero():
		state.Level = kernel.CompactAttention
		state.Title = kernel.NewText("The off-site copy failed")
		state.Detail = last()
	case s.Health == domain.HealthStale:
		state.Level = kernel.CompactAttention
		state.Title = kernel.NewText("Backup older than expected")
		state.Detail = last()
	default:
		state.Level = kernel.CompactOK
		state.Title = kernel.NewText("Backups all good")
		state.Detail = last()
	}
	return state
}
