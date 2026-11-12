package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/modules/services/usecase"
	"trisse-home/internal/shared/sqlutil"
)

// DiscoveryStore implements usecase.DiscoveryStore on SQLite (one row).
type DiscoveryStore struct{ db *sql.DB }

// NewDiscoveryStore returns a store backed by db.
func NewDiscoveryStore(db *sql.DB) *DiscoveryStore { return &DiscoveryStore{db: db} }

// DiscoverySettings returns the saved settings or the defaults.
func (s *DiscoveryStore) DiscoverySettings(ctx context.Context) (domain.DiscoverySettings, error) {
	var out domain.DiscoverySettings
	var enabled int
	var rule string
	err := s.db.QueryRowContext(ctx, `SELECT enabled, interval_minutes, rule FROM services_discovery WHERE id = 1`).
		Scan(&enabled, &out.IntervalMinutes, &rule)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultDiscoverySettings(), nil
	}
	if err != nil {
		return domain.DiscoverySettings{}, fmt.Errorf("select discovery settings: %w", err)
	}
	out.Enabled, out.Rule = enabled == 1, domain.MergeRule(rule)
	return out, nil
}

// SaveDiscoverySettings stores the settings and keeps the last result.
func (s *DiscoveryStore) SaveDiscoverySettings(ctx context.Context, d domain.DiscoverySettings) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO services_discovery (id, enabled, interval_minutes, rule) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled, interval_minutes = excluded.interval_minutes, rule = excluded.rule`,
		sqlutil.BoolToInt(d.Enabled), d.IntervalMinutes, string(d.Rule))
	if err != nil {
		return fmt.Errorf("save discovery settings: %w", err)
	}
	return nil
}

// LastSync returns the result of the last sync (zero before the first).
func (s *DiscoveryStore) LastSync(ctx context.Context) (usecase.SyncResult, error) {
	var r usecase.SyncResult
	var at string
	err := s.db.QueryRowContext(ctx, `SELECT last_sync_at, last_error, last_created, last_updated, last_hidden, last_shown, last_skipped
		FROM services_discovery WHERE id = 1`).Scan(&at, &r.Error, &r.Created, &r.Updated, &r.Hidden, &r.Shown, &r.Skipped)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.SyncResult{}, nil
	}
	if err != nil {
		return usecase.SyncResult{}, fmt.Errorf("select last sync: %w", err)
	}
	if at != "" {
		if r.At, err = sqlutil.ParseTime(at); err != nil {
			return usecase.SyncResult{}, err
		}
	}
	return r, nil
}

// SaveLastSync stores the result of a sync; settings default until saved.
func (s *DiscoveryStore) SaveLastSync(ctx context.Context, r usecase.SyncResult) error {
	def := domain.DefaultDiscoverySettings()
	_, err := s.db.ExecContext(ctx, `INSERT INTO services_discovery (id, interval_minutes, rule, last_sync_at, last_error,
		last_created, last_updated, last_hidden, last_shown, last_skipped) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET last_sync_at = excluded.last_sync_at, last_error = excluded.last_error,
		last_created = excluded.last_created, last_updated = excluded.last_updated, last_hidden = excluded.last_hidden,
		last_shown = excluded.last_shown, last_skipped = excluded.last_skipped`,
		def.IntervalMinutes, string(def.Rule), formatOptionalTime(r.At), r.Error, r.Created, r.Updated, r.Hidden, r.Shown, r.Skipped)
	if err != nil {
		return fmt.Errorf("save last sync: %w", err)
	}
	return nil
}

func formatOptionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return sqlutil.FormatTime(t)
}
