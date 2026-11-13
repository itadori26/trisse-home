package usecase

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"trisse-home/internal/modules/habits/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	habits map[string]domain.Habit
	done   map[string]map[domain.Day]bool
}

func newRepo() *memRepo {
	return &memRepo{habits: map[string]domain.Habit{}, done: map[string]map[domain.Day]bool{}}
}

func (m *memRepo) List(context.Context) ([]domain.Habit, error) {
	out := make([]domain.Habit, 0, len(m.habits))
	for _, h := range m.habits {
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (m *memRepo) Get(_ context.Context, id string) (domain.Habit, error) {
	h, ok := m.habits[id]
	if !ok {
		return domain.Habit{}, domain.ErrNotFound
	}
	return h, nil
}

func (m *memRepo) Save(_ context.Context, h domain.Habit) error { m.habits[h.ID] = h; return nil }

func (m *memRepo) Delete(_ context.Context, id string) error {
	delete(m.habits, id)
	delete(m.done, id)
	return nil
}

func (m *memRepo) Completions(_ context.Context, since domain.Day) (map[string]map[domain.Day]bool, error) {
	out := map[string]map[domain.Day]bool{}
	for id, days := range m.done {
		for d, ok := range days {
			if ok && !d.Before(since) {
				if out[id] == nil {
					out[id] = map[domain.Day]bool{}
				}
				out[id][d] = true
			}
		}
	}
	return out, nil
}

func (m *memRepo) SetCompletion(_ context.Context, id string, day domain.Day, done bool) error {
	if m.done[id] == nil {
		m.done[id] = map[domain.Day]bool{}
	}
	if done {
		m.done[id][day] = true
	} else {
		delete(m.done[id], day)
	}
	return nil
}

type clk struct{ t time.Time }

func (c clk) Now() time.Time { return c.t }

type seq struct{ n int }

func (s *seq) NewID() string { s.n++; return string(rune('a' + s.n - 1)) }

func TestHabits(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	// 00:30 in Berlin on the 13th is still the 12th in UTC: the user's day counts.
	berlin := time.FixedZone("CET", 3600)
	u := NewHabits(Deps{Repo: repo, Clock: clk{time.Date(2026, 11, 12, 23, 30, 0, 0, time.UTC)}, IDs: &seq{}, Location: berlin})

	if _, err := u.Create(ctx, Input{Name: "", Days: nil}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("empty name: %v", err)
	}
	read, err := u.Create(ctx, Input{Name: "Read", Emoji: "📖"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := u.Create(ctx, Input{Name: "Gym", Days: []int{1, 3, 5}}); err != nil {
		t.Fatal(err)
	}

	v, err := u.Toggle(ctx, read.Habit.ID, "2026-11-13") // today in Berlin
	if err != nil || !v.Stats.DoneToday || v.Stats.Current != 1 {
		t.Fatalf("toggle today: %+v %v", v.Stats, err)
	}
	if v, _ = u.Toggle(ctx, read.Habit.ID, "2026-11-12"); v.Stats.Current != 2 {
		t.Fatalf("yesterday: %+v", v.Stats)
	}
	if v, _ = u.Toggle(ctx, read.Habit.ID, "2026-11-13"); v.Stats.DoneToday || v.Stats.Current != 1 {
		t.Fatalf("undo today: %+v", v.Stats)
	}
	for _, d := range []string{"2026-11-11", "2026-11-14", "nope"} {
		if _, err := u.Toggle(ctx, read.Habit.ID, d); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", d, err)
		}
	}

	list, _ := u.List(ctx)
	if len(list) != 2 || list[0].Habit.Name != "Read" || list[1].Habit.Position != 1 || len(list[0].Stats.Heatmap) != domain.HeatmapDays {
		t.Fatalf("list %+v", list)
	}
	if _, err := u.Update(ctx, read.Habit.ID, Input{Name: "Read 20 pages", Days: []int{9}}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("bad weekday: %v", err)
	}
	if v, err := u.Update(ctx, read.Habit.ID, Input{Name: "Read 20 pages"}); err != nil || v.Habit.Name != "Read 20 pages" {
		t.Fatalf("update: %+v %v", v, err)
	}
	if err := u.Delete(ctx, read.Habit.ID); err != nil {
		t.Fatal(err)
	}
	if err := u.Delete(ctx, read.Habit.ID); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}
