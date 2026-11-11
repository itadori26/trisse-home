// Package gateway stores the waste settings and loads ICS subscriptions.
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
	"io"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/shared/ical"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

const maxICSBytes = 4 << 20

// SettingsRepository stores the settings.
type SettingsRepository struct{ db *sql.DB }

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB) *SettingsRepository { return &SettingsRepository{db: db} }

type ruleRow struct {
	Match string `json:"match"`
	Label string `json:"label"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// Load implements usecase.SettingsRepository.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var raw string
	err := r.db.QueryRowContext(ctx, `SELECT url, rules FROM waste_settings WHERE id = 1`).Scan(&s.URL, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	var rows []ruleRow
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return domain.Settings{}, fmt.Errorf("decode rules: %w", err)
	}
	s.Rules = make([]domain.Rule, 0, len(rows))
	for _, x := range rows {
		s.Rules = append(s.Rules, domain.Rule(x))
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	rows := make([]ruleRow, 0, len(s.Rules))
	for _, x := range s.Rules {
		rows = append(rows, ruleRow(x))
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode rules: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO waste_settings (id, url, rules) VALUES (1, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, rules = excluded.rules`, s.URL, string(raw))
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

// ICSSource implements usecase.CollectionSource.
type ICSSource struct {
	client *http.Client
	loc    *time.Location
}

// NewICSSource returns a source; floating times are interpreted in loc.
func NewICSSource(client *http.Client, loc *time.Location) *ICSSource {
	return &ICSSource{client: client, loc: loc}
}

// Collections implements usecase.CollectionSource.
func (s *ICSSource) Collections(ctx context.Context, url string, from, to time.Time) ([]domain.Collection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/calendar, */*;q=0.5")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server answered %s", res.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxICSBytes))
	if err != nil {
		return nil, fmt.Errorf("read calendar: %w", err)
	}
	events, err := ical.Parse(string(raw), s.loc)
	if err != nil {
		return nil, err
	}
	var out []domain.Collection
	for _, e := range events {
		for _, occ := range ical.Expand(e, from, to) {
			date := occ.Start
			if !occ.AllDay {
				date = date.In(s.loc)
			}
			out = append(out, domain.Collection{Date: domain.Date(date), Summary: occ.Summary})
		}
	}
	return out, nil
}
