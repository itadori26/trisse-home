package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Thursday, 12 November 2026.
var today = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

func day(offset int) Day { return today.AddDate(0, 0, offset) }

func doneOn(offsets ...int) map[Day]bool {
	m := map[Day]bool{}
	for _, o := range offsets {
		m[day(o)] = true
	}
	return m
}

func habit(t *testing.T, days ...int) Habit {
	t.Helper()
	s, err := NewSchedule(days)
	if err != nil {
		t.Fatal(err)
	}
	h, err := New("1", "Read", "📖", s, 0, day(-60))
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestScheduleAndValidation(t *testing.T) {
	if _, err := NewSchedule([]int{7}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("weekday 7: %v", err)
	}
	all, _ := NewSchedule([]int{0, 1, 2, 3, 4, 5, 6, 1})
	if !all.Daily() {
		t.Fatal("seven weekdays are daily")
	}
	weekdays, _ := NewSchedule([]int{5, 1, 1})
	if got := weekdays.Days(); len(got) != 2 || got[0] != 1 || got[1] != 5 {
		t.Fatalf("days %v", got)
	}
	if _, err := New("1", " ", "", Schedule{}, 0, today); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("empty name accepted")
	}
	if _, err := New("1", "x", "too long emoji", Schedule{}, 0, today); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("long emoji accepted")
	}
}

func TestStreaks(t *testing.T) {
	cases := []struct {
		name          string
		days          []int // schedule
		done          map[Day]bool
		current, best int
	}{
		{"nothing done", nil, doneOn(), 0, 0},
		{"today open does not break", nil, doneOn(-1, -2, -3), 3, 3},
		{"today done counts", nil, doneOn(0, -1, -2), 3, 3},
		{"gap breaks", nil, doneOn(0, -1, -3, -4, -5, -6), 2, 4},
		{"yesterday missed", nil, doneOn(-2, -3), 0, 2},
		// Mon, Wed, Fri: Thu (today) and Tue are not due and do not break it.
		{"days off the schedule", []int{1, 3, 5}, doneOn(-1, -3, -6, -8), 4, 4},
		{"a missed due day breaks", []int{1, 3, 5}, doneOn(-1, -6, -8), 1, 2},
		{"done on a free day is ignored", []int{1, 3, 5}, doneOn(-1, -2, -3), 2, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := ComputeStats(habit(t, c.days...), c.done, today)
			if st.Current != c.current || st.Best != c.best {
				t.Fatalf("current %d best %d, want %d %d", st.Current, st.Best, c.current, c.best)
			}
		})
	}
}

func TestHeatmap(t *testing.T) {
	st := ComputeStats(habit(t, 1, 3, 5), doneOn(0, -1, -27, -28), today)
	if len(st.Heatmap) != HeatmapDays || !st.Heatmap[HeatmapDays-1].Day.Equal(today) || !st.Heatmap[0].Day.Equal(day(-27)) {
		t.Fatalf("range %v … %v", st.Heatmap[0].Day, st.Heatmap[len(st.Heatmap)-1].Day)
	}
	last := st.Heatmap[HeatmapDays-1]
	if !last.Done || last.Due || !st.DoneToday || !st.Heatmap[HeatmapDays-2].Due {
		t.Fatalf("cells %+v", st.Heatmap[HeatmapDays-2:])
	}
}

func TestDayOfAndEditable(t *testing.T) {
	// 23:30 in Berlin is still the 12th there, although it is the 12th 22:30 UTC.
	berlin := time.FixedZone("CET", 3600)
	if got := DayOf(time.Date(2026, 11, 12, 23, 30, 0, 0, berlin)); !got.Equal(today) {
		t.Fatalf("day %v", got)
	}
	if EditableDay(today, today) != nil || EditableDay(day(-1), today) != nil {
		t.Fatal("today and yesterday must be editable")
	}
	if !errors.Is(EditableDay(day(-2), today), kernel.ErrInvalid) || !errors.Is(EditableDay(day(1), today), kernel.ErrInvalid) {
		t.Fatal("other days must not be editable")
	}
}
