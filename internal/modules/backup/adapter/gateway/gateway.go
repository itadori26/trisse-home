package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/modules/backup/domain"
)

// Snapshotter creates consistent copies of the live SQLite database.
type Snapshotter struct{ db *sql.DB }

// NewSnapshotter returns a snapshotter for db.
func NewSnapshotter(db *sql.DB) *Snapshotter { return &Snapshotter{db: db} }

// SnapshotTo implements usecase.Snapshotter. VACUUM INTO writes a compact,
// transactionally consistent copy while the application keeps running.
func (s *Snapshotter) SnapshotTo(ctx context.Context, path string) error {
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, path); err != nil {
		return fmt.Errorf("vacuum into %s: %w", filepath.Base(path), err)
	}
	return os.Chmod(path, 0o600)
}

// DirStore keeps backups as files in one directory.
type DirStore struct{ dir string }

// NewDirStore returns a store in dir, creating it if necessary.
func NewDirStore(dir string) (*DirStore, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}
	return &DirStore{dir: dir}, nil
}

// Path implements usecase.Store.
func (s *DirStore) Path(name string) string { return filepath.Join(s.dir, name) }

// List implements usecase.Store; files are ordered oldest first.
func (s *DirStore) List(context.Context) ([]domain.File, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read backup directory: %w", err)
	}
	var out []domain.File
	for _, e := range entries {
		if e.IsDir() || !domain.ValidName(e.Name()) {
			continue
		}
		f, err := s.stat(e.Name())
		if err != nil {
			continue
		}
		out = append(out, f)
	}
	slices.SortFunc(out, func(a, b domain.File) int { return a.CreatedAt.Compare(b.CreatedAt) })
	return out, nil
}

func (s *DirStore) stat(name string) (domain.File, error) {
	info, err := os.Stat(s.Path(name))
	if err != nil {
		return domain.File{}, err
	}
	stamp := strings.TrimSuffix(strings.TrimPrefix(name, "trisse-home-"), ".db")
	created, err := time.Parse("20060102-150405", stamp)
	if err != nil {
		created = info.ModTime().UTC()
	}
	return domain.File{Name: name, Size: info.Size(), CreatedAt: created}, nil
}

// Open implements usecase.Store. Only validated names are accepted.
func (s *DirStore) Open(name string) (io.ReadCloser, domain.File, error) {
	if !domain.ValidName(name) {
		return nil, domain.File{}, domain.ErrBackupNotFound
	}
	f, err := s.stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, domain.File{}, domain.ErrBackupNotFound
	}
	if err != nil {
		return nil, domain.File{}, fmt.Errorf("stat backup: %w", err)
	}
	rc, err := os.Open(s.Path(name))
	if err != nil {
		return nil, domain.File{}, fmt.Errorf("open backup: %w", err)
	}
	return rc, f, nil
}

// Remove implements usecase.Store.
func (s *DirStore) Remove(name string) error {
	if !domain.ValidName(name) {
		return domain.ErrBackupNotFound
	}
	return os.Remove(s.Path(name))
}
