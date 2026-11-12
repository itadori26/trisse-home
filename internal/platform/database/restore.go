package database

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Restore file suffixes next to the database (#46).
const (
	pendingSuffix = ".restore"
	beforeSuffix  = ".before-restore"
	stagePrefix   = ".restore-upload-"
)

// ErrTooLarge is returned when a staged restore exceeds the size limit.
var ErrTooLarge = errors.New("the backup file is too large")

// Restorer stages a database file for a restore and schedules the swap,
// which ApplyPendingRestore performs on the next start before the database
// is opened. Staged files live next to the database so the final rename is
// atomic on the same file system.
type Restorer struct {
	dbPath   string
	maxBytes int64
}

// NewRestorer returns a restorer for the database at dbPath.
func NewRestorer(dbPath string, maxBytes int64) *Restorer {
	return &Restorer{dbPath: dbPath, maxBytes: maxBytes}
}

// Stage copies r into a private temporary file and returns its path.
func (r *Restorer) Stage(ctx context.Context, src io.Reader) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(r.dbPath), stagePrefix+"*")
	if err != nil {
		return "", fmt.Errorf("create staging file: %w", err)
	}
	path := f.Name()
	fail := func(err error) (string, error) {
		_ = f.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := f.Chmod(0o600); err != nil {
		return fail(fmt.Errorf("chmod staging file: %w", err))
	}
	n, err := io.Copy(f, io.LimitReader(readerWithContext{ctx, src}, r.maxBytes+1))
	if err != nil {
		return fail(fmt.Errorf("write staging file: %w", err))
	}
	if n > r.maxBytes {
		return fail(ErrTooLarge)
	}
	if err := f.Sync(); err != nil {
		return fail(fmt.Errorf("sync staging file: %w", err))
	}
	if err := f.Close(); err != nil {
		return fail(fmt.Errorf("close staging file: %w", err))
	}
	return path, nil
}

// Discard removes a staged file (only files created by Stage).
func (r *Restorer) Discard(path string) {
	if r.owns(path) {
		_ = os.Remove(path)
	}
}

func (r *Restorer) owns(path string) bool {
	return filepath.Dir(path) == filepath.Dir(r.dbPath) && strings.HasPrefix(filepath.Base(path), stagePrefix)
}

// Schedule marks a staged file to replace the database on the next start.
func (r *Restorer) Schedule(path string) error {
	if !r.owns(path) {
		return errors.New("not a staged restore file")
	}
	if err := os.Rename(path, r.dbPath+pendingSuffix); err != nil {
		return fmt.Errorf("schedule restore: %w", err)
	}
	return nil
}

// ApplyPendingRestore swaps in a scheduled restore before the database is
// opened. The previous database (with its WAL files) is kept as
// "<db>.before-restore"; it reports whether a restore was applied.
func ApplyPendingRestore(dbPath string) (bool, error) {
	pending := dbPath + pendingSuffix
	if _, err := os.Stat(pending); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("check pending restore: %w", err)
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		src := dbPath + suffix
		dst := dbPath + beforeSuffix + suffix
		_ = os.Remove(dst)
		if err := os.Rename(src, dst); err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("keep previous database: %w", err)
		}
	}
	if err := os.Rename(pending, dbPath); err != nil {
		return false, fmt.Errorf("apply restore: %w", err)
	}
	return true, nil
}

type readerWithContext struct {
	ctx context.Context
	r   io.Reader
}

func (rc readerWithContext) Read(p []byte) (int, error) {
	if err := rc.ctx.Err(); err != nil {
		return 0, err
	}
	return rc.r.Read(p)
}
