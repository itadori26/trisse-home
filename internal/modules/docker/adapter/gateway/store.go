package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// UpdateStore implements usecase.UpdateCache and
// usecase.UpdateSettingsRepository in SQLite.
type UpdateStore struct {
	db     *sql.DB
	cipher Cipher
}

// NewUpdateStore returns a store backed by db.
func NewUpdateStore(db *sql.DB, cipher Cipher) *UpdateStore {
	return &UpdateStore{db: db, cipher: cipher}
}

// List implements usecase.UpdateCache.
func (s *UpdateStore) List(ctx context.Context) ([]usecase.ImageUpdate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT ref, local_digest, status, newer_tag, release_url, checked_at, error FROM docker_image_updates`)
	if err != nil {
		return nil, fmt.Errorf("query image updates: %w", err)
	}
	defer rows.Close()
	var out []usecase.ImageUpdate
	for rows.Next() {
		var u usecase.ImageUpdate
		var status, at string
		if err := rows.Scan(&u.Ref, &u.LocalDigest, &status, &u.NewerTag, &u.ReleaseURL, &at, &u.Error); err != nil {
			return nil, fmt.Errorf("scan image update: %w", err)
		}
		u.Status = domain.UpdateStatus(status)
		if u.CheckedAt, err = sqlutil.ParseTime(at); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// Put implements usecase.UpdateCache.
func (s *UpdateStore) Put(ctx context.Context, u usecase.ImageUpdate) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO docker_image_updates (ref, local_digest, status, newer_tag, release_url, checked_at, error)
		VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT (ref) DO UPDATE SET local_digest = excluded.local_digest, status = excluded.status,
		newer_tag = excluded.newer_tag, release_url = excluded.release_url, checked_at = excluded.checked_at, error = excluded.error`,
		u.Ref, u.LocalDigest, string(u.Status), u.NewerTag, u.ReleaseURL, sqlutil.FormatTime(u.CheckedAt), u.Error)
	if err != nil {
		return fmt.Errorf("upsert image update: %w", err)
	}
	return nil
}

// Retain implements usecase.UpdateCache.
func (s *UpdateStore) Retain(ctx context.Context, refs []string) error {
	if len(refs) == 0 {
		_, err := s.db.ExecContext(ctx, `DELETE FROM docker_image_updates`)
		return err
	}
	args := make([]any, len(refs))
	for i, r := range refs {
		args[i] = r
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(refs)), ",")
	_, err := s.db.ExecContext(ctx, `DELETE FROM docker_image_updates WHERE ref NOT IN (`+placeholders+`)`, args...) // #nosec G202 -- only placeholders
	return err
}

// Load implements usecase.UpdateSettingsRepository (defaults: off).
func (s *UpdateStore) Load(ctx context.Context) (usecase.UpdateSettings, error) {
	var out usecase.UpdateSettings
	err := s.db.QueryRowContext(ctx, `SELECT enabled, semver_tags FROM docker_update_settings WHERE id = 1`).Scan(&out.Enabled, &out.SemverTags)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return usecase.UpdateSettings{}, fmt.Errorf("query update settings: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT host, username, password_encrypted FROM docker_registry_credentials ORDER BY host`)
	if err != nil {
		return usecase.UpdateSettings{}, fmt.Errorf("query registry credentials: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c usecase.RegistryCredential
		var enc string
		if err := rows.Scan(&c.Host, &c.Username, &enc); err != nil {
			return usecase.UpdateSettings{}, fmt.Errorf("scan registry credential: %w", err)
		}
		if c.Password, err = s.cipher.Decrypt(enc); err != nil {
			return usecase.UpdateSettings{}, fmt.Errorf("decrypt registry password (was SECRET_KEY changed?): %w", err)
		}
		out.Credentials = append(out.Credentials, c)
	}
	return out, rows.Err()
}

// Save implements usecase.UpdateSettingsRepository.
func (s *UpdateStore) Save(ctx context.Context, set usecase.UpdateSettings) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO docker_update_settings (id, enabled, semver_tags) VALUES (1, ?, ?)
		ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled, semver_tags = excluded.semver_tags`, set.Enabled, set.SemverTags); err != nil {
		return fmt.Errorf("upsert update settings: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM docker_registry_credentials`); err != nil {
		return fmt.Errorf("clear registry credentials: %w", err)
	}
	for _, c := range set.Credentials {
		enc, err := s.cipher.Encrypt(c.Password)
		if err != nil {
			return fmt.Errorf("encrypt registry password: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO docker_registry_credentials (host, username, password_encrypted) VALUES (?, ?, ?)`,
			c.Host, c.Username, enc); err != nil {
			return fmt.Errorf("insert registry credential: %w", err)
		}
	}
	return tx.Commit()
}
