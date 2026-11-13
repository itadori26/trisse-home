// Package gateway stores the fuel settings in SQLite and reads prices from
// Tankerkönig.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/fuel/domain"
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

// SettingsRepository stores the settings with an encrypted API key.
type SettingsRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB, cipher Cipher) *SettingsRepository {
	return &SettingsRepository{db: db, cipher: cipher}
}

type favouriteRow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Load implements usecase.SettingsRepository.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var enc, fuel, favs string
	err := r.db.QueryRowContext(ctx, `SELECT api_key_encrypted, lat, lon, place, radius_km, fuel, favourites FROM fuel_settings WHERE id = 1`).
		Scan(&enc, &s.Lat, &s.Lon, &s.Place, &s.RadiusKm, &fuel, &favs)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.APIKey, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt API key (was SECRET_KEY changed?): %w", err)
	}
	var rows []favouriteRow
	if err := json.Unmarshal([]byte(favs), &rows); err != nil {
		return domain.Settings{}, fmt.Errorf("decode favourites: %w", err)
	}
	for _, f := range rows {
		s.Favourites = append(s.Favourites, domain.Favourite(f))
	}
	s.Fuel = domain.FuelType(fuel)
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.APIKey)
	if err != nil {
		return fmt.Errorf("encrypt API key: %w", err)
	}
	rows := make([]favouriteRow, 0, len(s.Favourites))
	for _, f := range s.Favourites {
		rows = append(rows, favouriteRow(f))
	}
	favs, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode favourites: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO fuel_settings (id, api_key_encrypted, lat, lon, place, radius_km, fuel, favourites)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET api_key_encrypted = excluded.api_key_encrypted, lat = excluded.lat, lon = excluded.lon,
			place = excluded.place, radius_km = excluded.radius_km, fuel = excluded.fuel, favourites = excluded.favourites`,
		enc, s.Lat, s.Lon, s.Place, s.RadiusKm, string(s.Fuel), string(favs))
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
