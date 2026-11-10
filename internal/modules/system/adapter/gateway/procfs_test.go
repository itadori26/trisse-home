package gateway

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"trisse-home/internal/modules/system/domain"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestHostReader(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	proc, sys := filepath.Join(root, "proc"), filepath.Join(root, "sys")
	writeFile(t, filepath.Join(proc, "stat"), "cpu  100 0 50 800 50 0 0 0 0 0\ncpu0 50 0 25 400 25 0 0 0 0 0\ncpu1 50 0 25 400 25 0 0 0 0 0\nintr 1\n")
	writeFile(t, filepath.Join(proc, "loadavg"), "0.52 0.41 0.30 1/234 5678\n")
	writeFile(t, filepath.Join(proc, "meminfo"), "MemTotal:        8000 kB\nMemFree:  100 kB\nMemAvailable:    2000 kB\nSwapTotal: 1000 kB\nSwapFree: 750 kB\n")
	writeFile(t, filepath.Join(proc, "uptime"), "3600.50 7000.00\n")
	writeFile(t, filepath.Join(sys, "class", "thermal", "thermal_zone0", "temp"), "48000\n")
	writeFile(t, filepath.Join(sys, "class", "thermal", "thermal_zone1", "temp"), "52500\n")

	h := NewHost(proc, sys, []DiskMount{{Label: "root", Path: root}, {Label: "missing", Path: "/does/not/exist"}})
	h.statfs = func(p string) (domain.Usage, error) {
		if p != root {
			return domain.Usage{}, os.ErrNotExist
		}
		return domain.Usage{Total: 100, Used: 40}, nil
	}

	cpu, err := h.CPUTimes(ctx)
	if err != nil || cpu.Total != 1000 || cpu.Idle != 850 {
		t.Fatalf("cpu %+v %v", cpu, err)
	}
	if n, _ := h.CPUCount(ctx); n != 2 {
		t.Fatalf("cpu count %d", n)
	}
	if l, _ := h.Load(ctx); l.Five != 0.41 {
		t.Fatalf("load %+v", l)
	}
	mem, swap, _ := h.Memory(ctx)
	if mem.Percent() != 75 || swap.Used != 250*1024 {
		t.Fatalf("mem %+v swap %+v", mem, swap)
	}
	if up, _ := h.Uptime(ctx); up.Seconds() != 3600.5 {
		t.Fatalf("uptime %v", up)
	}
	if d, _ := h.Disks(ctx); len(d) != 1 || d[0].Label != "root" {
		t.Fatalf("disks %+v", d)
	}
	if temp, _ := h.Temperature(ctx); temp != 52.5 {
		t.Fatalf("temperature %v", temp)
	}
}

func TestParseDiskMounts(t *testing.T) {
	got := ParseDiskMounts([]string{"Root=/host/rootfs", "/data"})
	if got[0] != (DiskMount{Label: "Root", Path: "/host/rootfs"}) || got[1] != (DiskMount{Label: "/data", Path: "/data"}) {
		t.Fatalf("unexpected %+v", got)
	}
}
