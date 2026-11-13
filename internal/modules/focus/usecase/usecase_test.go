package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/focus/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ sessions []domain.Session }

func (m *memRepo) Save(_ context.Context, s domain.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *memRepo) Since(_ context.Context, t time.Time) ([]domain.Session, error) {
	var out []domain.Session
	for _, s := range m.sessions {
		if !s.EndedAt.Before(t) {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *memRepo) DeleteBefore(_ context.Context, t time.Time) error {
	var keep []domain.Session
	for _, s := range m.sessions {
		if !s.EndedAt.Before(t) {
			keep = append(keep, s)
		}
	}
	m.sessions = keep
	return nil
}

type clk struct{ t time.Time }

func (c clk) Now() time.Time { return c.t }

type ids struct{}

func (ids) NewID() string { return "id" }

func TestRecordAndStats(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	repo := &memRepo{}
	u := NewFocus(Deps{Repo: repo, Clock: clk{now}, IDs: ids{}, Location: time.UTC})
	if _, err := u.RecordSession(ctx, 0, now); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid session accepted: %v", err)
	}
	for _, ago := range []time.Duration{time.Minute, time.Hour} {
		if _, err := u.RecordSession(ctx, 25, now.Add(-ago)); err != nil {
			t.Fatal(err)
		}
	}
	// An old session from before the week is ignored by the statistics.
	repo.sessions = append(repo.sessions, domain.Session{Minutes: 25, EndedAt: now.AddDate(0, 0, -10)})
	week, err := u.WeeklyStats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(week) != 7 || week[6].Sessions != 2 || week[6].Minutes != 50 {
		t.Fatalf("%+v", week)
	}
	repo.sessions = append(repo.sessions, domain.Session{Minutes: 25, EndedAt: now.AddDate(0, 0, -100)})
	if err := u.Prune(ctx); err != nil || len(repo.sessions) != 3 {
		t.Fatalf("prune: %v %d", err, len(repo.sessions))
	}
}
