// Package gateway stores the game deal settings in SQLite and reads the
// Steam and Epic Games stores.
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

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are part of data exports; a Steam ID is not a secret.
var TransferTables = []sqlutil.Table{{Name: "gamedeals_settings"}}

// SettingsRepository implements usecase.SettingsRepository.
type SettingsRepository struct{ db *sql.DB }

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB) *SettingsRepository { return &SettingsRepository{db: db} }

// Load returns the saved settings or the defaults (everything off).
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var epic int
	err := r.db.QueryRowContext(ctx, `SELECT steam_id, country, min_discount, epic FROM gamedeals_settings WHERE id = 1`).
		Scan(&s.SteamID, &s.Country, &s.MinDiscount, &epic)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	s.Epic = epic == 1
	return s, nil
}

// Save stores the settings.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO gamedeals_settings (id, steam_id, country, min_discount, epic) VALUES (1, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET steam_id = excluded.steam_id, country = excluded.country,
			min_discount = excluded.min_discount, epic = excluded.epic`,
		s.SteamID, s.Country, s.MinDiscount, sqlutil.BoolToInt(s.Epic))
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
