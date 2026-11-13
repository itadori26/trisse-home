package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" http://immich:2283/ ", " key ")
	if err != nil || s.URL != "http://immich:2283" || s.APIKey != "key" {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	for _, c := range [][2]string{{"immich", "k"}, {"http://x", " "}} {
		if _, err := NewSettings(c[0], c[1]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%v: expected a validation error, got %v", c, err)
		}
	}
}

func TestValidAssetIDAndDisk(t *testing.T) {
	if !ValidAssetID("3f2b6c1e-0a4d-4e5f-9a8b-1c2d3e4f5a6b") || ValidAssetID("../server") || ValidAssetID("3F2B6C1E-0A4D-4E5F-9A8B-1C2D3E4F5A6B") {
		t.Fatal("asset IDs")
	}
	if (LibraryStats{DiskUsed: 1, DiskTotal: 4}).DiskPercent() != 25 || (LibraryStats{}).DiskPercent() != 0 {
		t.Fatal("disk percent")
	}
}
