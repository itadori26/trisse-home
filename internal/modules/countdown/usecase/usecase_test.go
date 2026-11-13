package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/countdown/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ items map[string]domain.Countdown }

func (m *memRepo) List(context.Context) ([]domain.Countdown, error) {
	out := make([]domain.Countdown, 0, len(m.items))
	for _, c := range m.items {
		out = append(out, c)
	}
	return out, nil
}

func (m *memRepo) Get(_ context.Context, id string) (domain.Countdown, error) {
	c, ok := m.items[id]
	if !ok {
		return domain.Countdown{}, domain.ErrNotFound
	}
	return c, nil
}

func (m *memRepo) Save(_ context.Context, c domain.Countdown) error { m.items[c.ID] = c; return nil }
func (m *memRepo) Delete(_ context.Context, id string) error        { delete(m.items, id); return nil }

type clk struct{ t time.Time }

func (c clk) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return string(rune('a' + s.n - 1)) }

func newUC(now time.Time) (Countdowns, *memRepo) {
	repo := &memRepo{items: map[string]domain.Countdown{}}
	return NewCountdowns(Deps{Repo: repo, Clock: clk{now}, IDs: &seqIDs{}, Location: time.UTC}), repo
}

func TestCreateListSortsAndHidesPast(t *testing.T) {
	ctx := context.Background()
	u, _ := newUC(time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC))
	for _, in := range []domain.Input{
		{Title: "Holiday", Date: "2026-12-24"},
		{Title: "Exam", Date: "2026-11-20", Time: "09:00"},
		{Title: "Yesterday", Date: "2026-11-11"},  // ended at midnight today
		{Title: "Last month", Date: "2026-10-01"}, // long past
		{Title: "Deadline", Date: "2026-11-20", Time: "08:00"},
	} {
		if _, err := u.Create(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	titles := func(grace int) []string {
		list, err := u.List(ctx, grace)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, c := range list {
			out = append(out, c.Title)
		}
		return out
	}
	if got := titles(1); len(got) != 4 || got[0] != "Yesterday" || got[1] != "Deadline" || got[2] != "Exam" || got[3] != "Holiday" {
		t.Fatalf("grace 1: %v", got)
	}
	if got := titles(0); len(got) != 3 || got[0] != "Deadline" {
		t.Fatalf("grace 0: %v", got)
	}
	if got := titles(-5); len(got) != 3 {
		t.Fatalf("negative grace is clamped: %v", got)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	u, repo := newUC(time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC))
	c, err := u.Create(ctx, domain.Input{Title: "Exam", Date: "2026-12-01"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := u.Update(ctx, c.ID, domain.Input{Title: "", Date: "2026-12-01"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid update: %v", err)
	}
	if _, err := u.Update(ctx, c.ID, domain.Input{Title: "Exam 2", Emoji: "📚", Date: "2026-12-02", Time: "10:00"}); err != nil {
		t.Fatal(err)
	}
	if got := repo.items[c.ID]; got.Title != "Exam 2" || got.Time() != "10:00" || !got.CreatedAt.Equal(c.CreatedAt) {
		t.Fatalf("%+v", got)
	}
	if _, err := u.Update(ctx, "nope", domain.Input{Title: "X", Date: "2026-12-01"}); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unknown id: %v", err)
	}
	if err := u.Delete(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if err := u.Delete(ctx, c.ID); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}
