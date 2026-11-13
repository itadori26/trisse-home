// Package gateway stores the Proxmox settings in SQLite and talks to the
// Proxmox VE API.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/proxmox/domain"
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

// SettingsRepository stores the settings with an encrypted token secret.
type SettingsRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB, cipher Cipher) *SettingsRepository {
	return &SettingsRepository{db: db, cipher: cipher}
}

// Load implements usecase.SettingsRepository.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var enc string
	err := r.db.QueryRowContext(ctx, `SELECT url, token_id, secret_encrypted, fingerprint FROM proxmox_settings WHERE id = 1`).
		Scan(&s.URL, &s.TokenID, &enc, &s.Fingerprint)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.Secret, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt token secret (was SECRET_KEY changed?): %w", err)
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.Secret)
	if err != nil {
		return fmt.Errorf("encrypt token secret: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO proxmox_settings (id, url, token_id, secret_encrypted, fingerprint) VALUES (1, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, token_id = excluded.token_id,
		secret_encrypted = excluded.secret_encrypted, fingerprint = excluded.fingerprint`,
		s.URL, s.TokenID, enc, s.Fingerprint)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
