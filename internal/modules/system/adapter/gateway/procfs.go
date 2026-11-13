package gateway

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"trisse-home/internal/modules/system/domain"
)

// DiskMount is a file system to report: Label is shown, Path is statted.
type DiskMount struct {
	Label string
	Path  string
}

// ParseDiskMounts parses "label=path" or "path" entries.
func ParseDiskMounts(entries []string) []DiskMount {
	out := make([]DiskMount, 0, len(entries))
	for _, e := range entries {
		label, path, ok := strings.Cut(e, "=")
		if !ok {
			label, path = e, e
		}
		out = append(out, DiskMount{Label: strings.TrimSpace(label), Path: strings.TrimSpace(path)})
	}
	return out
}

// Host implements usecase.HostReader on top of procfs and sysfs.
type Host struct {
	procPath string
	sysPath  string
	disks    []DiskMount
	statfs   func(path string) (domain.Usage, error)
}

// NewHost returns a reader for the given proc and sys roots.
func NewHost(procPath, sysPath string, disks []DiskMount) *Host {
	return &Host{procPath: procPath, sysPath: sysPath, disks: disks, statfs: statfs}
}

func (h *Host) read(name string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(h.procPath, name)) //nolint:gosec // name is one of a few constant procfs file names
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return b, nil
}

// CPUTimes parses the aggregate cpu line of /proc/stat.
func (h *Host) CPUTimes(context.Context) (domain.CPUTimes, error) {
	b, err := h.read("stat")
	if err != nil {
		return domain.CPUTimes{}, err
	}
	line, _, _ := bytes.Cut(b, []byte("\n"))
	fields := strings.Fields(string(line))
	if len(fields) < 5 || fields[0] != "cpu" {
		return domain.CPUTimes{}, fmt.Errorf("unexpected /proc/stat format")
	}
	var t domain.CPUTimes
	for i, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return domain.CPUTimes{}, fmt.Errorf("parse /proc/stat: %w", err)
		}
		// guest and guest_nice (fields 9, 10) are already part of user/nice.
		if i < 8 {
			t.Total += v
		}
		if i == 3 || i == 4 { // idle and iowait
			t.Idle += v
		}
	}
	return t, nil
}

// CPUCount counts the cpuN lines of /proc/stat.
func (h *Host) CPUCount(context.Context) (int, error) {
	b, err := h.read("stat")
	if err != nil {
		return 0, err
	}
	n := 0
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		if line := s.Text(); strings.HasPrefix(line, "cpu") && len(line) > 3 && line[3] >= '0' && line[3] <= '9' {
			n++
		}
	}
	return n, nil
}

// Load parses /proc/loadavg.
func (h *Host) Load(context.Context) (domain.Load, error) {
	b, err := h.read("loadavg")
	if err != nil {
		return domain.Load{}, err
	}
	f := strings.Fields(string(b))
	if len(f) < 3 {
		return domain.Load{}, fmt.Errorf("unexpected /proc/loadavg format")
	}
	var l domain.Load
	for i, dst := range []*float64{&l.One, &l.Five, &l.Fifteen} {
		if *dst, err = strconv.ParseFloat(f[i], 64); err != nil {
			return domain.Load{}, fmt.Errorf("parse /proc/loadavg: %w", err)
		}
	}
	return l, nil
}

// Memory parses /proc/meminfo.
func (h *Host) Memory(context.Context) (domain.Usage, domain.Usage, error) {
	b, err := h.read("meminfo")
	if err != nil {
		return domain.Usage{}, domain.Usage{}, err
	}
	values := map[string]uint64{}
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		key, rest, ok := strings.Cut(s.Text(), ":")
		if !ok {
			continue
		}
		f := strings.Fields(rest)
		if len(f) == 0 {
			continue
		}
		if v, err := strconv.ParseUint(f[0], 10, 64); err == nil {
			values[key] = v
		}
	}
	mem := domain.MemoryFromMeminfo(values["MemTotal"], values["MemAvailable"])
	swap := domain.Usage{Total: values["SwapTotal"] * 1024, Used: (values["SwapTotal"] - min(values["SwapFree"], values["SwapTotal"])) * 1024}
	return mem, swap, nil
}

// Uptime parses /proc/uptime.
func (h *Host) Uptime(context.Context) (time.Duration, error) {
	b, err := h.read("uptime")
	if err != nil {
		return 0, err
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return 0, fmt.Errorf("unexpected /proc/uptime format")
	}
	secs, err := strconv.ParseFloat(f[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse /proc/uptime: %w", err)
	}
	return time.Duration(secs * float64(time.Second)), nil
}

// maxHostnameLength is the longest valid DNS host name.
const maxHostnameLength = 253

// Hostname reads sys/kernel/hostname of the host's procfs (not the
// container's own name). It returns "" when the file is missing or empty;
// control characters are dropped and the length is capped, because the name is
// shown in the UI.
func (h *Host) Hostname(context.Context) string {
	b, err := h.read(filepath.Join("sys", "kernel", "hostname"))
	if err != nil {
		return ""
	}
	name := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(string(b)))
	if len(name) > maxHostnameLength {
		name = name[:maxHostnameLength]
	}
	return name
}

// Disks stats every configured mount.
func (h *Host) Disks(context.Context) ([]domain.Disk, error) {
	out := make([]domain.Disk, 0, len(h.disks))
	for _, d := range h.disks {
		u, err := h.statfs(d.Path)
		if err != nil {
			continue // a missing mount should not hide the other metrics
		}
		out = append(out, domain.Disk{Label: d.Label, Usage: u})
	}
	return out, nil
}

// milliDegrees converts sysfs temperature readings.
const milliDegrees = 1000.0

// Temperature returns the highest reading of all thermal zones.
func (h *Host) Temperature(context.Context) (float64, error) {
	paths, _ := filepath.Glob(filepath.Join(h.sysPath, "class", "thermal", "thermal_zone*", "temp"))
	hwmon, _ := filepath.Glob(filepath.Join(h.sysPath, "class", "hwmon", "hwmon*", "temp*_input"))
	var highest float64
	for _, p := range append(paths, hwmon...) {
		b, err := os.ReadFile(p) //nolint:gosec // paths come from a fixed glob below the configured sysfs root
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
		if err != nil {
			continue
		}
		highest = max(highest, v/milliDegrees)
	}
	if highest == 0 {
		return 0, fmt.Errorf("no temperature sensor found")
	}
	return highest, nil
}
