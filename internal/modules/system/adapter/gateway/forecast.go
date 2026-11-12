package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/system/domain"
)

// AggregateStore implements usecase.AggregateRepository in SQLite.
type AggregateStore struct{ db *sql.DB }

// NewAggregateStore returns a store backed by db.
func NewAggregateStore(db *sql.DB) *AggregateStore { return &AggregateStore{db: db} }

// SQLite integers are signed; usage in bytes always fits into int64.
func toInt(v uint64) int64 { return int64(min(v, 1<<62)) } // #nosec G115 -- clamped

func toUint(v int64) uint64 { return uint64(max(v, 0)) } // #nosec G115 -- clamped

// Get implements usecase.AggregateRepository.
func (s *AggregateStore) Get(ctx context.Context, mount, day string) (domain.UsageAggregate, bool, error) {
	var minU, maxU, last, total int64
	err := s.db.QueryRowContext(ctx, `SELECT min_used, max_used, last_used, total FROM system_disk_daily WHERE mount = ? AND day = ?`,
		mount, day).Scan(&minU, &maxU, &last, &total)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.UsageAggregate{}, false, nil
	}
	if err != nil {
		return domain.UsageAggregate{}, false, fmt.Errorf("query aggregate: %w", err)
	}
	return domain.UsageAggregate{Mount: mount, Day: day, MinUsed: toUint(minU), MaxUsed: toUint(maxU), Last: toUint(last), Total: toUint(total)}, true, nil
}

// Put implements usecase.AggregateRepository.
func (s *AggregateStore) Put(ctx context.Context, a domain.UsageAggregate) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO system_disk_daily (mount, day, min_used, max_used, last_used, total) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (mount, day) DO UPDATE SET min_used = excluded.min_used, max_used = excluded.max_used,
		last_used = excluded.last_used, total = excluded.total`,
		a.Mount, a.Day, toInt(a.MinUsed), toInt(a.MaxUsed), toInt(a.Last), toInt(a.Total))
	if err != nil {
		return fmt.Errorf("upsert aggregate: %w", err)
	}
	return nil
}

// Since implements usecase.AggregateRepository.
func (s *AggregateStore) Since(ctx context.Context, day string) ([]domain.UsageAggregate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT mount, day, min_used, max_used, last_used, total FROM system_disk_daily
		WHERE day >= ? ORDER BY mount, day`, day)
	if err != nil {
		return nil, fmt.Errorf("query aggregates: %w", err)
	}
	defer rows.Close()
	var out []domain.UsageAggregate
	for rows.Next() {
		var a domain.UsageAggregate
		var minU, maxU, last, total int64
		if err := rows.Scan(&a.Mount, &a.Day, &minU, &maxU, &last, &total); err != nil {
			return nil, fmt.Errorf("scan aggregate: %w", err)
		}
		a.MinUsed, a.MaxUsed, a.Last, a.Total = toUint(minU), toUint(maxU), toUint(last), toUint(total)
		out = append(out, a)
	}
	return out, rows.Err()
}

// PruneBefore implements usecase.AggregateRepository.
func (s *AggregateStore) PruneBefore(ctx context.Context, day string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM system_disk_daily WHERE day < ?`, day)
	return err
}

// ForecastSettings implements usecase.AggregateRepository.
func (s *AggregateStore) ForecastSettings(ctx context.Context) (domain.ForecastSettings, bool, error) {
	var fs domain.ForecastSettings
	err := s.db.QueryRowContext(ctx, `SELECT window_days, alert_days FROM system_forecast_settings WHERE id = 1`).Scan(&fs.WindowDays, &fs.AlertDays)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ForecastSettings{}, false, nil
	}
	if err != nil {
		return domain.ForecastSettings{}, false, fmt.Errorf("query forecast settings: %w", err)
	}
	return fs, true, nil
}

// SaveForecastSettings implements usecase.AggregateRepository.
func (s *AggregateStore) SaveForecastSettings(ctx context.Context, fs domain.ForecastSettings) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO system_forecast_settings (id, window_days, alert_days) VALUES (1, ?, ?)
		ON CONFLICT (id) DO UPDATE SET window_days = excluded.window_days, alert_days = excluded.alert_days`, fs.WindowDays, fs.AlertDays)
	return err
}
