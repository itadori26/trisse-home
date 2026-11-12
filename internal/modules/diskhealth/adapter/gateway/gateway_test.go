package gateway

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"trisse-home/internal/modules/diskhealth/domain"
)

func TestReports(t *testing.T) {
	reports, err := NewFileSource("testdata").Reports(context.Background())
	if err != nil || len(reports) != 2 {
		t.Fatalf("%+v %v", reports, err)
	}
	nvme, sda := reports[0], reports[1]
	if nvme.Protocol != domain.ProtocolNVMe || nvme.NVMe == nil || nvme.NVMe.AvailableSpare != 100 || nvme.CapacityB != 1000204886016 || *nvme.Temperature != 41 {
		t.Fatalf("nvme: %+v", nvme)
	}
	if sda.Protocol != domain.ProtocolATA || sda.Model != "WDC WD40EFRX-68N32N0" || !*sda.Passed || *sda.PowerOnH != 33712 || len(sda.Attributes) != 5 || sda.At.Unix() != 1794470400 {
		t.Fatalf("sda: %+v", sda)
	}
	if sda.Attributes[1].ID != 5 || sda.Attributes[1].Raw != 2 {
		t.Fatalf("attributes: %+v", sda.Attributes)
	}
}

func TestReportsWithoutExport(t *testing.T) {
	for _, dir := range []string{"", filepath.Join(t.TempDir(), "missing"), t.TempDir()} {
		if _, err := NewFileSource(dir).Reports(context.Background()); !errors.Is(err, domain.ErrNoExport) {
			t.Fatalf("%q: %v", dir, err)
		}
	}
}

func TestBrokenFileKeepsOthers(t *testing.T) {
	dir := t.TempDir()
	good, _ := os.ReadFile("testdata/sda.json")
	_ = os.WriteFile(filepath.Join(dir, "sda.json"), good, 0o600)
	_ = os.WriteFile(filepath.Join(dir, "sdb.json"), []byte("{not json"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, ".export.tmp.json"), []byte("{}"), 0o600)
	reports, err := NewFileSource(dir).Reports(context.Background())
	if err != nil || len(reports) != 2 || reports[1].Device != "sdb" || reports[1].Passed != nil {
		t.Fatalf("%+v %v", reports, err)
	}
}
