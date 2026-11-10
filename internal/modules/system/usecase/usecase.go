// Package usecase samples host metrics in the background and serves the
// latest snapshot.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/shared/kernel"
)

// HostReader reads raw metrics of the host (output port).
type HostReader interface {
	CPUTimes(ctx context.Context) (domain.CPUTimes, error)
	CPUCount(ctx context.Context) (int, error)
	Load(ctx context.Context) (domain.Load, error)
	Memory(ctx context.Context) (mem, swap domain.Usage, err error)
	Uptime(ctx context.Context) (time.Duration, error)
	Disks(ctx context.Context) ([]domain.Disk, error)
	// Temperature returns the highest sensor reading in °C, or 0 if unknown.
	Temperature(ctx context.Context) (float64, error)
}

// Monitor is the input port of the system module.
type Monitor interface {
	// Sample takes a new snapshot; it is run by the background job.
	Sample(ctx context.Context) error
	// Current returns the latest snapshot, sampling once if none exists.
	Current(ctx context.Context) (SnapshotView, error)
}

// UsageView is a used/total pair in bytes with its percentage.
type UsageView struct {
	Total, Used uint64
	Percent     float64
}

// DiskView is the usage of one file system.
type DiskView struct {
	Label string
	Usage UsageView
}

// SnapshotView is the host metrics snapshot at the use case boundary.
type SnapshotView struct {
	CPUPercent   float64
	CPUCount     int
	Load1, Load5 float64
	Load15       float64
	Memory, Swap UsageView
	Disks        []DiskView
	Uptime       time.Duration
	Temperature  float64
	TakenAt      time.Time
}

func usageView(u domain.Usage) UsageView {
	return UsageView{Total: u.Total, Used: u.Used, Percent: u.Percent()}
}

func toView(s domain.Snapshot) SnapshotView {
	v := SnapshotView{
		CPUPercent: s.CPUPercent, CPUCount: s.CPUCount, Load1: s.Load.One, Load5: s.Load.Five, Load15: s.Load.Fifteen,
		Memory: usageView(s.Memory), Swap: usageView(s.Swap), Disks: make([]DiskView, 0, len(s.Disks)),
		Uptime: s.Uptime, Temperature: s.Temperature, TakenAt: s.TakenAt,
	}
	for _, d := range s.Disks {
		v.Disks = append(v.Disks, DiskView{Label: d.Label, Usage: usageView(d.Usage)})
	}
	return v
}

type monitor struct {
	host  HostReader
	clock kernel.Clock
	mu    sync.RWMutex
	prev  *domain.CPUTimes
	last  *domain.Snapshot
}

// NewMonitor returns the system monitor.
func NewMonitor(host HostReader, clock kernel.Clock) Monitor {
	return &monitor{host: host, clock: clock}
}

// Sample implements Monitor. CPU usage needs two counter readings, so the
// first sample reports 0 % until the next one.
func (m *monitor) Sample(ctx context.Context) error {
	cpu, err := m.host.CPUTimes(ctx)
	if err != nil {
		return fmt.Errorf("read cpu times: %w", err)
	}
	snap := domain.Snapshot{TakenAt: m.clock.Now()}
	m.mu.RLock()
	if m.prev != nil {
		snap.CPUPercent = domain.CPUUsage(*m.prev, cpu)
	}
	m.mu.RUnlock()
	if snap.CPUCount, err = m.host.CPUCount(ctx); err != nil {
		return fmt.Errorf("read cpu count: %w", err)
	}
	if snap.Load, err = m.host.Load(ctx); err != nil {
		return fmt.Errorf("read load: %w", err)
	}
	if snap.Memory, snap.Swap, err = m.host.Memory(ctx); err != nil {
		return fmt.Errorf("read memory: %w", err)
	}
	if snap.Uptime, err = m.host.Uptime(ctx); err != nil {
		return fmt.Errorf("read uptime: %w", err)
	}
	if snap.Disks, err = m.host.Disks(ctx); err != nil {
		return fmt.Errorf("read disks: %w", err)
	}
	// Temperature sensors are optional (VMs often have none).
	snap.Temperature, _ = m.host.Temperature(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.prev, m.last = &cpu, &snap
	return nil
}

// Current implements Monitor.
func (m *monitor) Current(ctx context.Context) (SnapshotView, error) {
	m.mu.RLock()
	last := m.last
	m.mu.RUnlock()
	if last != nil {
		return toView(*last), nil
	}
	if err := m.Sample(ctx); err != nil {
		return SnapshotView{}, kernel.NewPublicError(kernel.ErrUnavailable, "Host metrics are not available. Are /proc and /sys mounted?")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return toView(*m.last), nil
}
