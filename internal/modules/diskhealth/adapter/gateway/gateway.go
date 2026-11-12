// Package gateway reads the smartctl JSON files written by
// scripts/smart-export.sh.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"trisse-home/internal/modules/diskhealth/domain"
)

// Limits protecting against a broken or hostile export directory.
const (
	maxFiles = 64
	maxBytes = 4 << 20
)

// FileSource implements usecase.SmartReportSource for an export directory
// with one "<disk>.json" (output of smartctl --json -a) per disk.
type FileSource struct{ dir string }

// NewFileSource reads exports from dir (empty means not configured).
func NewFileSource(dir string) *FileSource { return &FileSource{dir: dir} }

// Reports implements usecase.SmartReportSource.
func (f *FileSource) Reports(_ context.Context) ([]domain.Report, error) {
	if f.dir == "" {
		return nil, domain.ErrNoExport
	}
	entries, err := os.ReadDir(f.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, domain.ErrNoExport
	}
	if err != nil {
		return nil, fmt.Errorf("read export directory: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), ".json") && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return nil, domain.ErrNoExport
	}
	sort.Strings(names)
	if len(names) > maxFiles {
		names = names[:maxFiles]
	}
	out := make([]domain.Report, 0, len(names))
	for _, n := range names {
		r, err := readReport(filepath.Join(f.dir, n))
		if err != nil {
			// One unreadable file must not hide the other disks.
			r = domain.Report{Device: strings.TrimSuffix(n, ".json"), Protocol: domain.ProtocolOther, At: time.Unix(0, 0)}
		}
		out = append(out, r)
	}
	return out, nil
}

func readReport(path string) (domain.Report, error) {
	file, err := os.Open(path) // #nosec G304 -- path is inside the configured export directory
	if err != nil {
		return domain.Report{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return domain.Report{}, err
	}
	r, err := Parse(io.LimitReader(file, maxBytes))
	if err != nil {
		return domain.Report{}, err
	}
	if r.At.IsZero() {
		r.At = info.ModTime()
	}
	if r.Device == "" {
		r.Device = strings.TrimSuffix(filepath.Base(path), ".json")
	}
	return r, nil
}

// smartctl is the subset of `smartctl --json -a` that is evaluated.
type smartctl struct {
	Device struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	ModelName    string `json:"model_name"`
	UserCapacity struct {
		Bytes int64 `json:"bytes"`
	} `json:"user_capacity"`
	NVMeCapacity int64 `json:"nvme_total_capacity"`
	LocalTime    struct {
		TimeT int64 `json:"time_t"`
	} `json:"local_time"`
	SmartStatus *struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	Temperature *struct {
		Current int `json:"current"`
	} `json:"temperature"`
	PowerOnTime *struct {
		Hours int64 `json:"hours"`
	} `json:"power_on_time"`
	ATA *struct {
		Table []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Value      int    `json:"value"`
			Thresh     int    `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Raw        struct {
				Value int64 `json:"value"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
	NVMe *struct {
		CriticalWarning int   `json:"critical_warning"`
		AvailableSpare  int   `json:"available_spare"`
		SpareThreshold  int   `json:"available_spare_threshold"`
		PercentageUsed  int   `json:"percentage_used"`
		MediaErrors     int64 `json:"media_errors"`
	} `json:"nvme_smart_health_information_log"`
}

// Parse normalises the JSON output of `smartctl --json -a`.
func Parse(rd io.Reader) (domain.Report, error) {
	var s smartctl
	if err := json.NewDecoder(rd).Decode(&s); err != nil {
		return domain.Report{}, fmt.Errorf("decode smartctl json: %w", err)
	}
	r := domain.Report{Device: s.Device.Name, Model: strings.TrimSpace(s.ModelName), CapacityB: s.UserCapacity.Bytes}
	switch strings.ToUpper(s.Device.Protocol) {
	case "ATA":
		r.Protocol = domain.ProtocolATA
	case "NVME":
		r.Protocol = domain.ProtocolNVMe
	default:
		r.Protocol = domain.ProtocolOther
	}
	if r.CapacityB == 0 {
		r.CapacityB = s.NVMeCapacity
	}
	if s.LocalTime.TimeT > 0 {
		r.At = time.Unix(s.LocalTime.TimeT, 0)
	}
	if s.SmartStatus != nil {
		passed := s.SmartStatus.Passed
		r.Passed = &passed
	}
	if s.Temperature != nil && s.Temperature.Current > 0 {
		t := s.Temperature.Current
		r.Temperature = &t
	}
	if s.PowerOnTime != nil {
		h := s.PowerOnTime.Hours
		r.PowerOnH = &h
	}
	if s.ATA != nil {
		for _, a := range s.ATA.Table {
			r.Attributes = append(r.Attributes, domain.Attribute{
				ID: a.ID, Name: a.Name, Value: a.Value, Threshold: a.Thresh, Raw: a.Raw.Value,
				FailingNow: a.WhenFailed == "now",
			})
		}
	}
	if n := s.NVMe; n != nil {
		r.NVMe = &domain.NVMeHealth{CriticalWarning: n.CriticalWarning, AvailableSpare: n.AvailableSpare,
			SpareThreshold: n.SpareThreshold, PercentageUsed: n.PercentageUsed, MediaErrors: n.MediaErrors}
	}
	return r, nil
}
