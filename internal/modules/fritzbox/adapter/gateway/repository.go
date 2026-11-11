// Package gateway stores the encrypted FRITZ!Box settings and talks to the
// box via UPnP/IGD and TR-064 (SOAP with HTTP digest authentication).
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

	"trisse-home/internal/modules/fritzbox/domain"
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

// SettingsRepository stores the settings with an encrypted password.
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
	err := r.db.QueryRowContext(ctx, `SELECT url, username, password_encrypted FROM fritzbox_settings WHERE id = 1`).Scan(&s.URL, &s.Username, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if enc != "" {
		if s.Password, err = r.cipher.Decrypt(enc); err != nil {
			return domain.Settings{}, fmt.Errorf("decrypt password (was SECRET_KEY changed?): %w", err)
		}
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc := ""
	if s.Password != "" {
		var err error
		if enc, err = r.cipher.Encrypt(s.Password); err != nil {
			return fmt.Errorf("encrypt password: %w", err)
		}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO fritzbox_settings (id, url, username, password_encrypted) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, username = excluded.username, password_encrypted = excluded.password_encrypted`,
		s.URL, s.Username, enc)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
