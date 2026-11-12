package usecase

import (
	"context"
	"sort"
	"testing"
	"time"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/shared/kernel"
)

type memAggregates struct {
	rows     map[string]domain.UsageAggregate
	settings *domain.ForecastSettings
}

func (m *memAggregates) Get(_ context.Context, mount, day string) (domain.UsageAggregate, bool, error) {
	a, ok := m.rows[mount+"|"+day]
	return a, ok, nil
}
func (m *memAggregates) Put(_ context.Context, a domain.UsageAggregate) error {
	m.rows[a.Mount+"|"+a.Day] = a
	return nil
}
func (m *memAggregates) Since(_ context.Context, day string) ([]domain.UsageAggregate, error) {
	var out []domain.UsageAggregate
	for _, a := range m.rows {
		if a.Day >= day {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mount+out[i].Day < out[j].Mount+out[j].Day })
	return out, nil
}
func (m *memAggregates) PruneBefore(_ context.Context, day string) error {
	for k, a := range m.rows {
		if a.Day < day {
			delete(m.rows, k)
		}
	}
	return nil
}
func (m *memAggregates) ForecastSettings(context.Context) (domain.ForecastSettings, bool, error) {
	if m.settings == nil {
		return domain.ForecastSettings{}, false, nil
	}
	return *m.settings, true, nil
}
func (m *memAggregates) SaveForecastSettings(_ context.Context, s domain.ForecastSettings) error {
	m.settings = &s
	return nil
}

type fixedDisks struct{ used uint64 }

func (f *fixedDisks) Current(context.Context) (SnapshotView, error) {
	return SnapshotView{Disks: []DiskView{{Label: "System", Usage: UsageView{Used: f.used, Total: 100 << 30}}}}, nil
}

type observations struct{ got []kernel.Observation }

func (o *observations) Observe(_ context.Context, ob kernel.Observation) { o.got = append(o.got, ob) }

type dayClock struct{ t time.Time }

func (c *dayClock) Now() time.Time { return c.t }

func TestDiskForecasts(t *testing.T) {
	ctx := context.Background()
	repo := &memAggregates{rows: map[string]domain.UsageAggregate{}}
	disks := &fixedDisks{}
	obs := &observations{}
	clk := &dayClock{t: time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)}
	uc := NewDiskForecasts(disks, repo, obs, clk, time.UTC)

	// 2 GB per day on a 100 GB disk, starting at 70 GB: full in ~15 days at day 1, < 14 days later.
	for day := range 6 {
		disks.used = uint64(70+2*day) << 30
		if err := uc.Aggregate(ctx); err != nil {
			t.Fatal(err)
		}
		clk.t = clk.t.Add(24 * time.Hour)
	}
	clk.t = clk.t.Add(-24 * time.Hour)
	f, err := uc.Forecasts(ctx)
	if err != nil || f["System"].Status != domain.ForecastFilling || f["System"].DaysUntilFull < 9 || f["System"].DaysUntilFull > 11 {
		t.Fatalf("%+v %v", f, err)
	}
	last := obs.got[len(obs.got)-1]
	if last.Kind != "forecast" || last.Healthy || last.Detail == "" {
		t.Fatalf("observation: %+v", last)
	}
	if _, err := uc.SaveSettings(ctx, domain.ForecastSettings{WindowDays: 1, AlertDays: 5}); err == nil {
		t.Fatal("invalid window accepted")
	}
	if _, err := uc.SaveSettings(ctx, domain.ForecastSettings{WindowDays: 14, AlertDays: 5}); err != nil {
		t.Fatal(err)
	}
	_ = uc.Aggregate(ctx)
	if last := obs.got[len(obs.got)-1]; !last.Healthy {
		t.Fatal("10 days left is fine with a 5-day threshold")
	}
}

func TestAggregateFoldsSameDay(t *testing.T) {
	ctx := context.Background()
	repo := &memAggregates{rows: map[string]domain.UsageAggregate{}}
	disks := &fixedDisks{used: 50}
	clk := &dayClock{t: time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)}
	uc := NewDiskForecasts(disks, repo, kernel.DiscardObservations{}, clk, time.UTC)
	_ = uc.Aggregate(ctx)
	disks.used = 40
	_ = uc.Aggregate(ctx)
	disks.used = 45
	_ = uc.Aggregate(ctx)
	a := repo.rows["System|2026-11-01"]
	if a.MinUsed != 40 || a.MaxUsed != 50 || a.Last != 45 || len(repo.rows) != 1 {
		t.Fatalf("%+v", repo.rows)
	}
}
