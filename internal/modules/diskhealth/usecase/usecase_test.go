package usecase

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/diskhealth/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeSource struct {
	reports []domain.Report
	err     error
}

func (f fakeSource) Reports(context.Context) ([]domain.Report, error) { return f.reports, f.err }

type recorder struct{ got []kernel.Observation }

func (r *recorder) Observe(_ context.Context, o kernel.Observation) { r.got = append(r.got, o) }

type clock struct{ t time.Time }

func (c clock) Now() time.Time { return c.t }

func TestDiskHealth(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	ok, failed := true, false
	src := fakeSource{reports: []domain.Report{
		{Device: "/dev/sdb", At: now, Passed: &failed},
		{Device: "/dev/sda", At: now, Passed: &ok},
		{Device: "/dev/sdc", At: now.Add(-5 * time.Hour), Passed: &ok}, // stale
	}}
	rec := &recorder{}
	uc := NewDiskHealth(src, rec, clock{now}, 0)
	disks, err := uc.Disks(context.Background())
	if err != nil || len(disks) != 3 || disks[0].Device != "/dev/sda" || disks[1].Status != domain.StatusCritical || disks[2].Status != domain.StatusUnknown {
		t.Fatalf("%+v %v", disks, err)
	}
	if err := uc.Evaluate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(rec.got) != 2 || !rec.got[0].Healthy || rec.got[1].Healthy {
		t.Fatalf("observations: %+v", rec.got)
	}
}

func TestEvaluateWithoutExport(t *testing.T) {
	uc := NewDiskHealth(fakeSource{err: domain.ErrNoExport}, nil, clock{time.Now()}, 0)
	if err := uc.Evaluate(context.Background()); err != nil {
		t.Fatalf("missing export must not fail the job: %v", err)
	}
}
