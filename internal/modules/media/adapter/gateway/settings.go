// Package gateway stores the media server settings in SQLite and talks to
// Jellyfin and Plex.
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

	"trisse-home/internal/modules/media/domain"
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

// SettingsRepository stores the settings with an encrypted token.
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
	var provider, enc string
	err := r.db.QueryRowContext(ctx, `SELECT provider, url, token_encrypted FROM media_settings WHERE id = 1`).Scan(&provider, &s.URL, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.Token, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt token (was SECRET_KEY changed?): %w", err)
	}
	s.Provider = domain.Provider(provider)
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.Token)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO media_settings (id, provider, url, token_encrypted) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET provider = excluded.provider, url = excluded.url, token_encrypted = excluded.token_encrypted`,
		string(s.Provider), s.URL, enc)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
