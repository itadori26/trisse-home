package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/shared/kernel"
)

func TestStatusFollowsTheBackupOutcome(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 21, 3, 0, 0, 0, time.UTC)}
	store := &memStore{files: map[string]domain.File{}}
	snap := &fakeSnap{store: store, clock: c}
	policy, _ := domain.NewPolicy(24*time.Hour, 7)
	tracker := &Tracker{}
	uc := NewBackups(policy, snap, store, c, WithTracker(tracker))

	s, err := uc.Status(ctx)
	if err != nil || s.Health != domain.HealthStale || !s.LastBackupAt.IsZero() || s.Count != 0 {
		t.Fatalf("no backup yet: %+v %v", s, err)
	}

	if _, err := uc.BackupNow(ctx); err != nil {
		t.Fatal(err)
	}
	if s, _ = uc.Status(ctx); s.Health != domain.HealthOK || !s.LastBackupAt.Equal(c.t) || s.Count != 1 {
		t.Fatalf("after a backup: %+v", s)
	}

	// The next night's backup fails: red, with the reason.
	c.t = c.t.Add(24 * time.Hour)
	snap.err = errors.New("disk full")
	if _, err := uc.BackupNow(ctx); err == nil {
		t.Fatal("expected the failure")
	}
	if s, _ = uc.Status(ctx); s.Health != domain.HealthFailed || s.Error == "" || !s.FailedAt.Equal(c.t) {
		t.Fatalf("after a failure: %+v", s)
	}

	// A working backup turns it green again.
	c.t = c.t.Add(time.Hour)
	snap.err = nil
	if err := uc.RunScheduled(ctx); err != nil {
		t.Fatal(err)
	}
	if s, _ = uc.Status(ctx); s.Health != domain.HealthOK || s.Error != "" {
		t.Fatalf("after the retry: %+v", s)
	}

	// A failed off-site copy is amber until the next upload works.
	tracker.OffsiteResult(c.t, errors.New("401 Unauthorized"))
	if s, _ = uc.Status(ctx); s.Health != domain.HealthStale || s.OffsiteError == "" {
		t.Fatalf("off-site failure: %+v", s)
	}
	tracker.OffsiteResult(c.t, nil)
	if s, _ = uc.Status(ctx); s.Health != domain.HealthOK {
		t.Fatalf("off-site retry: %+v", s)
	}

	// Two days without a backup: older than expected.
	c.t = c.t.Add(48 * time.Hour)
	if s, _ = uc.Status(ctx); s.Health != domain.HealthStale {
		t.Fatalf("overdue: %+v", s)
	}
}

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	last := now.Add(-9 * time.Hour)
	tests := []struct {
		name  string
		s     StatusView
		level kernel.CompactLevel
		title string
	}{
		{"ok", StatusView{Health: domain.HealthOK, LastBackupAt: last}, kernel.CompactOK, "Backups all good"},
		{"failed", StatusView{Health: domain.HealthFailed, LastBackupAt: last, Outcome: domain.Outcome{FailedAt: now}}, kernel.CompactProblem, "The last backup failed"},
		{"none", StatusView{Health: domain.HealthStale}, kernel.CompactAttention, "No backup yet"},
		{"overdue", StatusView{Health: domain.HealthStale, LastBackupAt: last}, kernel.CompactAttention, "Backup older than expected"},
		{"off-site", StatusView{Health: domain.HealthStale, LastBackupAt: last, Outcome: domain.Outcome{OffsiteFailedAt: now}}, kernel.CompactAttention, "The off-site copy failed"},
	}
	for _, tt := range tests {
		got := CompactState(tt.s, now, time.UTC)
		if got.Level != tt.level || got.Title.String() != tt.title {
			t.Errorf("%s: %s %q", tt.name, got.Level, got.Title.String())
		}
	}
	if d := CompactState(tests[0].s, now, time.UTC).Detail.String(); d != "Last backup 2026-11-21 03:00" {
		t.Errorf("detail: %q", d)
	}
}
