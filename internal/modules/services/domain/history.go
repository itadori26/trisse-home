package domain

import "time"

// History limits.
const (
	HistoryRetention = 24 * time.Hour
	MaxChartPoints   = 96
)

// LatencyPoint is one stored health check result.
type LatencyPoint struct {
	ServiceID string
	At        time.Time
	LatencyMs float64
	Online    bool
}

// DownsampleLatency averages consecutive points into at most maxPoints
// buckets; a bucket is online only if all its checks were online.
func DownsampleLatency(points []LatencyPoint, maxPoints int) []LatencyPoint {
	if maxPoints <= 0 || len(points) <= maxPoints {
		return points
	}
	size := (len(points) + maxPoints - 1) / maxPoints
	out := make([]LatencyPoint, 0, maxPoints)
	for start := 0; start < len(points); start += size {
		end := min(start+size, len(points))
		b := LatencyPoint{ServiceID: points[start].ServiceID, At: points[end-1].At, Online: true}
		var sum float64
		for _, p := range points[start:end] {
			sum += p.LatencyMs
			b.Online = b.Online && p.Online
		}
		b.LatencyMs = sum / float64(end-start)
		out = append(out, b)
	}
	return out
}
