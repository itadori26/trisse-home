package domain

import (
	"testing"
	"time"
)

func TestHistoryRules(t *testing.T) {
	now := time.Unix(1000, 0)
	if !ShouldRecord(time.Time{}, now) || ShouldRecord(now, now.Add(30*time.Second)) || !ShouldRecord(now, now.Add(time.Minute)) {
		t.Fatal("ShouldRecord wrong")
	}
	snap := Snapshot{CPUPercent: 10, Memory: Usage{Total: 100, Used: 50}, Disks: []Disk{{Usage: Usage{Total: 10, Used: 2}}, {Usage: Usage{Total: 10, Used: 9}}}}
	if p := PointFrom(snap); p.Memory != 50 || p.Disk != 90 {
		t.Fatalf("PointFrom %+v", p)
	}
	var points []Point
	for i := range 10 {
		points = append(points, Point{At: now.Add(time.Duration(i) * time.Minute), CPU: float64(i)})
	}
	got := Downsample(points, 5)
	if len(got) != 5 || got[0].CPU != 0.5 || !got[4].At.Equal(points[9].At) {
		t.Fatalf("downsample %+v", got)
	}
	if len(Downsample(points, 20)) != 10 {
		t.Fatal("short series must stay unchanged")
	}
}
