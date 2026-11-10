package domain

import "testing"

func TestCPUUsage(t *testing.T) {
	tests := []struct {
		prev, cur CPUTimes
		want      float64
	}{
		{CPUTimes{Idle: 100, Total: 200}, CPUTimes{Idle: 150, Total: 300}, 50},
		{CPUTimes{Idle: 100, Total: 200}, CPUTimes{Idle: 200, Total: 300}, 0},
		{CPUTimes{Idle: 100, Total: 200}, CPUTimes{Idle: 100, Total: 200}, 0},
		{CPUTimes{Idle: 100, Total: 200}, CPUTimes{Idle: 100, Total: 300}, 100},
	}
	for _, tt := range tests {
		if got := CPUUsage(tt.prev, tt.cur); got != tt.want {
			t.Errorf("CPUUsage(%v,%v) = %v, want %v", tt.prev, tt.cur, got, tt.want)
		}
	}
}

func TestMemoryFromMeminfo(t *testing.T) {
	m := MemoryFromMeminfo(8000, 2000)
	if m.Total != 8000*1024 || m.Used != 6000*1024 || m.Percent() != 75 {
		t.Fatalf("unexpected %+v %.1f", m, m.Percent())
	}
	if MemoryFromMeminfo(100, 200).Used != 0 {
		t.Fatal("available > total must not underflow")
	}
	if (Usage{}).Percent() != 0 {
		t.Fatal("zero total must give 0 percent")
	}
}
