package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/focus/domain"
	"trisse-home/internal/shared/kernel"
)

// Focus is the input port of the module; every method is one use case
// (#58).
type Focus interface {
	// RecordSession stores a finished focus phase.
	RecordSession(ctx context.Context, minutes int, endedAt time.Time) (domain.Session, error)
	// WeeklyStats counts sessions per day for the last seven days.
	WeeklyStats(ctx context.Context) ([]domain.Day, error)
	// Prune deletes sessions older than the retention (job).
	Prune(ctx context.Context) error
}

// Repository persists sessions (output port, implemented in
// adapter/gateway).
type Repository interface {
	Save(ctx context.Context, s domain.Session) error
	Since(ctx context.Context, t time.Time) ([]domain.Session, error)
	DeleteBefore(ctx context.Context, t time.Time) error
}

// Retention is how long sessions are kept.
const Retention = 90 * 24 * time.Hour

// Deps bundles the output ports. Location is the user's time zone
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewFocus returns the focus timer use cases.
func NewFocus(d Deps) Focus {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

// RecordSession implements Focus.
func (u *uc) RecordSession(ctx context.Context, minutes int, endedAt time.Time) (domain.Session, error) {
	s, err := domain.NewSession(u.d.IDs.NewID(), minutes, endedAt, u.d.Clock.Now())
	if err != nil {
		return domain.Session{}, err
	}
	if err := u.d.Repo.Save(ctx, s); err != nil {
		return domain.Session{}, fmt.Errorf("save session: %w", err)
	}
	return s, nil
}

// WeeklyStats implements Focus.
func (u *uc) WeeklyStats(ctx context.Context) ([]domain.Day, error) {
	now := u.d.Clock.Now().In(u.d.Location)
	sessions, err := u.d.Repo.Since(ctx, domain.WeekStart(now))
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return domain.Week(sessions, now), nil
}

// Prune implements Focus.
func (u *uc) Prune(ctx context.Context) error {
	return u.d.Repo.DeleteBefore(ctx, u.d.Clock.Now().Add(-Retention))
}
