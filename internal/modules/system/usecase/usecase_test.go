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
	m := NewMonitor(host, clk{})
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
	m := NewMonitor(&fakeHost{err: errors.New("no /proc")}, clk{})
	if _, err := m.Current(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
