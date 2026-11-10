package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"trisse-home/internal/modules/calendar/domain"
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

// SettingsRepository stores the CalDAV settings with an encrypted password.
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
	var urls, username, enc string
	err := r.db.QueryRowContext(ctx, `SELECT calendar_urls, username, password_encrypted FROM calendar_settings WHERE id = 1`).Scan(&urls, &username, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	password, err := r.cipher.Decrypt(enc)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt password (was SECRET_KEY changed?): %w", err)
	}
	return domain.Settings{CalendarURLs: strings.Split(urls, "\n"), Username: username, Password: password}, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO calendar_settings (id, calendar_urls, username, password_encrypted) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET calendar_urls = excluded.calendar_urls, username = excluded.username,
			password_encrypted = excluded.password_encrypted`,
		strings.Join(s.CalendarURLs, "\n"), s.Username, enc)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
