package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/sqlutil"
)

// UptimeStore implements usecase.UptimeStore in SQLite.
type UptimeStore struct{ db *sql.DB }

// NewUptimeStore returns the aggregate and incident store.
func NewUptimeStore(db *sql.DB) *UptimeStore { return &UptimeStore{db: db} }

// AddCheck implements usecase.UptimeStore.
func (s *UptimeStore) AddCheck(ctx context.Context, serviceID, day string, online, maintenance bool) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO services_daily (service_id, day, checks, online, maintenance) VALUES (?, ?, 1, ?, ?)
		ON CONFLICT (service_id, day) DO UPDATE SET checks = checks + 1, online = online + excluded.online,
		maintenance = maintenance + excluded.maintenance`,
		serviceID, day, sqlutil.BoolToInt(online && !maintenance), sqlutil.BoolToInt(maintenance))
	if err != nil {
		return fmt.Errorf("upsert aggregate: %w", err)
	}
	return nil
}

// Aggregates implements usecase.UptimeStore.
func (s *UptimeStore) Aggregates(ctx context.Context, sinceDay string) ([]domain.DailyAggregate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT service_id, day, checks, online, maintenance FROM services_daily WHERE day >= ? ORDER BY day`, sinceDay)
	if err != nil {
		return nil, fmt.Errorf("query aggregates: %w", err)
	}
	defer rows.Close()
	var out []domain.DailyAggregate
	for rows.Next() {
		var a domain.DailyAggregate
		if err := rows.Scan(&a.ServiceID, &a.Day, &a.Checks, &a.Online, &a.Maintenance); err != nil {
			return nil, fmt.Errorf("scan aggregate: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// PruneAggregatesBefore implements usecase.UptimeStore.
func (s *UptimeStore) PruneAggregatesBefore(ctx context.Context, day string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM services_daily WHERE day < ?`, day)
	return err
}

const incidentColumns = `id, service_id, started_at, ended_at, cause`

func scanIncident(sc interface{ Scan(...any) error }) (domain.Incident, error) {
	var i domain.Incident
	var start string
	var end *string
	if err := sc.Scan(&i.ID, &i.ServiceID, &start, &end, &i.Cause); err != nil {
		return i, err
	}
	var err error
	if i.Start, err = sqlutil.ParseTime(start); err != nil {
		return i, err
	}
	i.End, err = sqlutil.ParseNullableTime(end)
	return i, err
}

// OpenIncident implements usecase.UptimeStore.
func (s *UptimeStore) OpenIncident(ctx context.Context, serviceID string) (domain.Incident, bool, error) {
	i, err := scanIncident(s.db.QueryRowContext(ctx, `SELECT `+incidentColumns+` FROM services_incidents
		WHERE service_id = ? AND ended_at IS NULL ORDER BY started_at DESC LIMIT 1`, serviceID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Incident{}, false, nil
	}
	if err != nil {
		return domain.Incident{}, false, fmt.Errorf("select incident: %w", err)
	}
	return i, true, nil
}

// SaveIncident implements usecase.UptimeStore (upsert).
func (s *UptimeStore) SaveIncident(ctx context.Context, i domain.Incident) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO services_incidents (`+incidentColumns+`) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET ended_at = excluded.ended_at, cause = excluded.cause`,
		i.ID, i.ServiceID, sqlutil.FormatTime(i.Start), sqlutil.FormatNullableTime(i.End), i.Cause)
	if err != nil {
		return fmt.Errorf("upsert incident: %w", err)
	}
	return nil
}

// Incidents implements usecase.UptimeStore.
func (s *UptimeStore) Incidents(ctx context.Context, serviceID string, limit int) ([]domain.Incident, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+incidentColumns+` FROM services_incidents
		WHERE (? = '' OR service_id = ?) ORDER BY started_at DESC LIMIT ?`, serviceID, serviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("query incidents: %w", err)
	}
	defer rows.Close()
	out := []domain.Incident{}
	for rows.Next() {
		i, err := scanIncident(rows)
		if err != nil {
			return nil, fmt.Errorf("scan incident: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// PruneIncidentsBefore implements usecase.UptimeStore (closed incidents only).
func (s *UptimeStore) PruneIncidentsBefore(ctx context.Context, t time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM services_incidents WHERE ended_at IS NOT NULL AND ended_at < ?`, sqlutil.FormatTime(t))
	return err
}
