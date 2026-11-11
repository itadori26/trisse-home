package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/shared/kernel"
)

// FileView is a backup at the use case boundary.
type FileView struct {
	Name      string
	Size      int64
	CreatedAt time.Time
}

// Backups is the input port of the module; every method is one use case.
type Backups interface {
	// RunScheduled takes a backup if one is due and prunes old ones (job).
	RunScheduled(ctx context.Context) error
	// BackupNow takes a backup immediately and prunes old ones.
	BackupNow(ctx context.Context) (FileView, error)
	List(ctx context.Context) ([]FileView, error)
	// Open returns the content of a backup for download.
	Open(ctx context.Context, name string) (io.ReadCloser, FileView, error)
}

// Snapshotter writes a consistent copy of the live database (output port).
type Snapshotter interface {
	SnapshotTo(ctx context.Context, path string) error
}

// Store manages backup files (output port).
type Store interface {
	// Path returns the absolute path for a new backup file name.
	Path(name string) string
	List(ctx context.Context) ([]domain.File, error)
	Open(name string) (io.ReadCloser, domain.File, error)
	Remove(name string) error
}

type backups struct {
	policy domain.Policy
	snap   Snapshotter
	store  Store
	clock  kernel.Clock
}

// NewBackups returns the backup use cases.
func NewBackups(policy domain.Policy, snap Snapshotter, store Store, clock kernel.Clock) Backups {
	return &backups{policy: policy, snap: snap, store: store, clock: clock}
}

func toView(f domain.File) FileView {
	return FileView{Name: f.Name, Size: f.Size, CreatedAt: f.CreatedAt}
}

// RunScheduled implements Backups.
func (uc *backups) RunScheduled(ctx context.Context) error {
	files, err := uc.store.List(ctx)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	var latest time.Time
	for _, f := range files {
		if f.CreatedAt.After(latest) {
			latest = f.CreatedAt
		}
	}
	if !uc.policy.Due(latest, uc.clock.Now()) {
		return nil
	}
	_, err = uc.BackupNow(ctx)
	return err
}

// BackupNow implements Backups.
func (uc *backups) BackupNow(ctx context.Context) (FileView, error) {
	now := uc.clock.Now()
	name := domain.FileName(now)
	if err := uc.snap.SnapshotTo(ctx, uc.store.Path(name)); err != nil {
		return FileView{}, fmt.Errorf("create backup: %w", err)
	}
	if err := uc.prune(ctx); err != nil {
		return FileView{}, err
	}
	rc, f, err := uc.store.Open(name)
	if err != nil {
		return FileView{}, err
	}
	_ = rc.Close()
	return toView(f), nil
}

func (uc *backups) prune(ctx context.Context) error {
	files, err := uc.store.List(ctx)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	for _, f := range uc.policy.ToPrune(files) {
		if err := uc.store.Remove(f.Name); err != nil {
			return fmt.Errorf("remove old backup: %w", err)
		}
	}
	return nil
}

// List implements Backups; newest first.
func (uc *backups) List(ctx context.Context) ([]FileView, error) {
	files, err := uc.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	out := make([]FileView, 0, len(files))
	for i := len(files) - 1; i >= 0; i-- {
		out = append(out, toView(files[i]))
	}
	return out, nil
}

// Open implements Backups. Names are validated so requests cannot escape
// the backup directory.
func (uc *backups) Open(_ context.Context, name string) (io.ReadCloser, FileView, error) {
	if !domain.ValidName(name) {
		return nil, FileView{}, domain.ErrBackupNotFound
	}
	rc, f, err := uc.store.Open(name)
	if err != nil {
		return nil, FileView{}, err
	}
	return rc, toView(f), nil
}
