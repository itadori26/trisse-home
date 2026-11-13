package domain

import "time"

// CPUTimes are cumulative jiffies from the aggregate "cpu" line of /proc/stat.
type CPUTimes struct {
	Idle  uint64
	Total uint64
}

// CPUUsage returns the busy share in percent between two samples. It returns
// 0 when the counters did not advance (e.g. identical samples).
func CPUUsage(prev, cur CPUTimes) float64 {
	if cur.Total <= prev.Total || cur.Idle < prev.Idle {
		return 0
	}
	total := float64(cur.Total - prev.Total)
	idle := float64(cur.Idle - prev.Idle)
	return clampPercent((total - idle) / total * 100)
}

// Usage is a used/total pair in bytes.
type Usage struct {
	Total uint64
	Used  uint64
}

// Percent returns the used share in percent.
func (u Usage) Percent() float64 {
	if u.Total == 0 {
		return 0
	}
	return clampPercent(float64(u.Used) / float64(u.Total) * 100)
}

// MemoryFromMeminfo derives used memory the way `free` does: total minus
// available, so page cache does not count as used.
func MemoryFromMeminfo(totalKiB, availableKiB uint64) Usage {
	if availableKiB > totalKiB {
		availableKiB = totalKiB
	}
	return Usage{Total: totalKiB * 1024, Used: (totalKiB - availableKiB) * 1024}
}

// Disk is the usage of one mounted file system.
type Disk struct {
	Label string
	Usage Usage
}

// Load holds the 1, 5 and 15 minute load averages.
type Load struct {
	One, Five, Fifteen float64
}

// Snapshot is a complete set of host metrics at one point in time.
type Snapshot struct {
	// Hostname of the machine; empty when it cannot be read (#123).
	Hostname    string
	CPUPercent  float64
	CPUCount    int
	Load        Load
	Memory      Usage
	Swap        Usage
	Disks       []Disk
	Uptime      time.Duration
	Temperature float64 // °C, 0 if unknown
	TakenAt     time.Time
}

func clampPercent(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 100:
		return 100
	default:
		return v
	}
}
