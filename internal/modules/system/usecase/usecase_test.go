package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/system/domain"
)

type fakeHost struct {
	cpu  []domain.CPUTimes
	call int
	err  error
}

func (f *fakeHost) CPUTimes(context.Context) (domain.CPUTimes, error) {
	if f.err != nil {
		return domain.CPUTimes{}, f.err
	}
	c := f.cpu[min(f.call, len(f.cpu)-1)]
	f.call++
	return c, nil
}
func (f *fakeHost) CPUCount(context.Context) (int, error)         { return 4, nil }
func (f *fakeHost) Load(context.Context) (domain.Load, error)     { return domain.Load{One: 0.5}, nil }
func (f *fakeHost) Uptime(context.Context) (time.Duration, error) { return time.Hour, nil }
func (f *fakeHost) Disks(context.Context) ([]domain.Disk, error) {
	return []domain.Disk{{Label: "/"}}, nil
}
func (f *fakeHost) Temperature(context.Context) (float64, error) { return 0, errors.New("no sensor") }
func (f *fakeHost) Memory(context.Context) (domain.Usage, domain.Usage, error) {
	return domain.Usage{Total: 100, Used: 50}, domain.Usage{}, nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestMonitor(t *testing.T) {
	ctx := context.Background()
	host := &fakeHost{cpu: []domain.CPUTimes{{Idle: 0, Total: 0}, {Idle: 75, Total: 100}}}
	m := NewMonitor(host, &memHistory{}, clk{})
	first, err := m.Current(ctx)
	if err != nil || first.CPUPercent != 0 || first.CPUCount != 4 {
		t.Fatalf("first snapshot: %+v %v", first, err)
	}
	if err := m.Sample(ctx); err != nil {
		t.Fatal(err)
	}
	second, _ := m.Current(ctx)
	if second.CPUPercent != 25 {
		t.Fatalf("cpu = %v, want 25", second.CPUPercent)
	}
}

func TestMonitorUnavailable(t *testing.T) {
	m := NewMonitor(&fakeHost{err: errors.New("no /proc")}, &memHistory{}, clk{})
	if _, err := m.Current(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

type memHistory struct{ points []domain.Point }

func (m *memHistory) Add(_ context.Context, p domain.Point) error {
	m.points = append(m.points, p)
	return nil
}
func (m *memHistory) Since(_ context.Context, t time.Time) ([]domain.Point, error) {
	var out []domain.Point
	for _, p := range m.points {
		if !p.At.Before(t) {
			out = append(out, p)
		}
	}
	return out, nil
}
func (m *memHistory) PruneBefore(_ context.Context, t time.Time) error {
	var keep []domain.Point
	for _, p := range m.points {
		if !p.At.Before(t) {
			keep = append(keep, p)
		}
	}
	m.points = keep
	return nil
}

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

func TestHistoryIsRecordedOncePerMinute(t *testing.T) {
	ctx := context.Background()
	c := &movingClock{t: time.Unix(1_000_000, 0)}
	hist := &memHistory{}
	m := NewMonitor(&fakeHost{cpu: []domain.CPUTimes{{Total: 1}, {Total: 2}}}, hist, c)
	for range 30 { // 30 samples, 5 s apart = 2.5 minutes
		if err := m.Sample(ctx); err != nil {
			t.Fatal(err)
		}
		c.t = c.t.Add(5 * time.Second)
	}
	if len(hist.points) != 3 { // first sample skipped, then one per minute
		t.Fatalf("expected 3 points, got %d", len(hist.points))
	}
	c.t = c.t.Add(25 * time.Hour)
	_ = m.Sample(ctx)
	if len(hist.points) != 1 {
		t.Fatalf("old points not pruned: %d", len(hist.points))
	}
	if got, _ := m.History(ctx, 1); len(got) != 1 {
		t.Fatalf("history %v", got)
	}
}
