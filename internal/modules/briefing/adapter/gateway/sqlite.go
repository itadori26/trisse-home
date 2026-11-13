// Package gateway stores the briefing settings in SQLite.
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
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/modules/briefing/usecase"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{{Name: "briefing_settings"}}

// Repository implements usecase.SettingsRepository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

type layoutRow struct {
	Kind    string `json:"kind"`
	Enabled bool   `json:"enabled"`
}

// Load implements usecase.SettingsRepository.
func (r *Repository) Load(ctx context.Context) (usecase.Settings, bool, error) {
	var layout, days string
	var on, skip bool
	var minute int
	var lat, lon float64
	var last *string
	err := r.db.QueryRowContext(ctx, `SELECT layout, push_on, push_minute, push_days, skip_empty, lat, lon, last_sent
		FROM briefing_settings WHERE id = 1`).Scan(&layout, &on, &minute, &days, &skip, &lat, &lon, &last)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Settings{}, false, nil
	}
	if err != nil {
		return usecase.Settings{}, false, fmt.Errorf("select settings: %w", err)
	}
	s := usecase.Settings{Lat: lat, Lon: lon, Push: domain.PushSchedule{Enabled: on, Minute: minute, SkipEmpty: skip}}
	var rows []layoutRow
	_ = json.Unmarshal([]byte(layout), &rows) // unreadable layouts fall back to the default
	for _, row := range rows {
		s.Layout = append(s.Layout, domain.LayoutEntry{Kind: domain.SectionKind(row.Kind), Enabled: row.Enabled})
	}
	for _, p := range strings.Split(days, ",") {
		if d, err := strconv.Atoi(p); err == nil && d >= 0 && d <= 6 {
			s.Push.Days = append(s.Push.Days, time.Weekday(d))
		}
	}
	if s.LastSent, err = sqlutil.ParseNullableTime(last); err != nil {
		return usecase.Settings{}, false, err
	}
	return s, true, nil
}

// Save implements usecase.SettingsRepository.
func (r *Repository) Save(ctx context.Context, s usecase.Settings) error {
	rows := make([]layoutRow, 0, len(s.Layout))
	for _, e := range s.Layout {
		rows = append(rows, layoutRow{Kind: string(e.Kind), Enabled: e.Enabled})
	}
	layout, _ := json.Marshal(rows) // plain structs, cannot fail
	days := make([]string, 0, len(s.Push.Days))
	for _, d := range s.Push.Days {
		days = append(days, strconv.Itoa(int(d)))
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO briefing_settings (id, layout, push_on, push_minute, push_days, skip_empty, lat, lon, last_sent)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET layout = excluded.layout, push_on = excluded.push_on, push_minute = excluded.push_minute,
			push_days = excluded.push_days, skip_empty = excluded.skip_empty, lat = excluded.lat, lon = excluded.lon,
			last_sent = excluded.last_sent`,
		string(layout), s.Push.Enabled, s.Push.Minute, strings.Join(days, ","), s.Push.SkipEmpty, s.Lat, s.Lon,
		sqlutil.FormatNullableTime(s.LastSent))
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
