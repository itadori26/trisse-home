package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// EventRepository implements usecase.SecurityEventRepository in SQLite.
type EventRepository struct{ db *sql.DB }

// NewEventRepository returns the security event store.
func NewEventRepository(db *sql.DB) *EventRepository { return &EventRepository{db: db} }

// Add implements usecase.SecurityEventRepository.
func (r *EventRepository) Add(ctx context.Context, e domain.SecurityEvent) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO auth_security_events (id, at, type, account_id, ip, user_agent, detail)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, e.ID, sqlutil.FormatTime(e.At), string(e.Type), e.AccountID, e.Client.IP, e.Client.UserAgent, e.Detail)
	if err != nil {
		return fmt.Errorf("insert security event: %w", err)
	}
	return nil
}

// List implements usecase.SecurityEventRepository (newest first).
func (r *EventRepository) List(ctx context.Context, limit int) ([]domain.SecurityEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, at, type, account_id, ip, user_agent, detail
		FROM auth_security_events ORDER BY at DESC, id LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query security events: %w", err)
	}
	defer rows.Close()
	var out []domain.SecurityEvent
	for rows.Next() {
		var e domain.SecurityEvent
		var at, typ string
		if err := rows.Scan(&e.ID, &at, &typ, &e.AccountID, &e.Client.IP, &e.Client.UserAgent, &e.Detail); err != nil {
			return nil, fmt.Errorf("scan security event: %w", err)
		}
		e.Type = domain.EventType(typ)
		if e.At, err = sqlutil.ParseTime(at); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// KnownDevice implements usecase.SecurityEventRepository.
func (r *EventRepository) KnownDevice(ctx context.Context, accountID, userAgent string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM auth_security_events
		WHERE account_id = ? AND type = ? AND user_agent = ?`, accountID, string(domain.EventLoginSucceeded), userAgent).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("query known devices: %w", err)
	}
	return n > 0, nil
}

// DeleteBefore implements usecase.SecurityEventRepository.
func (r *EventRepository) DeleteBefore(ctx context.Context, t time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_security_events WHERE at < ?`, sqlutil.FormatTime(t))
	if err != nil {
		return 0, fmt.Errorf("delete security events: %w", err)
	}
	return res.RowsAffected()
}
