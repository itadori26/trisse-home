package domain

import "time"

// History limits.
const (
	// HistoryResolution is the minimum distance between stored points.
	HistoryResolution = time.Minute
	// HistoryRetention is how long points are kept.
	HistoryRetention = 24 * time.Hour
	// MaxChartPoints bounds the number of points sent to the browser.
	MaxChartPoints = 288
)

// Point is one stored sample of the host metrics in percent.
type Point struct {
	At     time.Time
	CPU    float64
	Memory float64
	Disk   float64
}

// PointFrom summarises a snapshot; Disk is the fullest file system.
func PointFrom(s Snapshot) Point {
	p := Point{At: s.TakenAt, CPU: s.CPUPercent, Memory: s.Memory.Percent()}
	for _, d := range s.Disks {
		p.Disk = max(p.Disk, d.Usage.Percent())
	}
	return p
}

// ShouldRecord reports whether a new point is due after last.
func ShouldRecord(last, now time.Time) bool {
	return last.IsZero() || now.Sub(last) >= HistoryResolution
}

// Downsample averages consecutive points into at most maxPoints buckets so
// charts stay light regardless of the time range.
func Downsample(points []Point, maxPoints int) []Point {
	if maxPoints <= 0 || len(points) <= maxPoints {
		return points
	}
	size := (len(points) + maxPoints - 1) / maxPoints
	out := make([]Point, 0, maxPoints)
	for start := 0; start < len(points); start += size {
		end := min(start+size, len(points))
		var sum Point
		for _, p := range points[start:end] {
			sum.CPU += p.CPU
			sum.Memory += p.Memory
			sum.Disk += p.Disk
		}
		n := float64(end - start)
		out = append(out, Point{At: points[end-1].At, CPU: sum.CPU / n, Memory: sum.Memory / n, Disk: sum.Disk / n})
	}
	return out
}
