package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/calendar/domain"
)

// SubscriptionRepository stores ICS subscriptions with encrypted URLs.
type SubscriptionRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSubscriptionRepository returns a repository backed by db.
func NewSubscriptionRepository(db *sql.DB, cipher Cipher) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, cipher: cipher}
}

const subscriptionColumns = `id, name, url_encrypted, color, refresh_minutes`

func (r *SubscriptionRepository) scan(row interface{ Scan(...any) error }) (domain.Subscription, error) {
	var s domain.Subscription
	var enc string
	var refresh int
	if err := row.Scan(&s.ID, &s.Name, &enc, &s.Color, &refresh); err != nil {
		return domain.Subscription{}, err
	}
	url, err := r.cipher.Decrypt(enc)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("decrypt subscription URL (was SECRET_KEY changed?): %w", err)
	}
	s.URL, s.RefreshInterval = url, time.Duration(refresh)*time.Minute
	return s, nil
}

// List implements usecase.SubscriptionRepository in creation order.
func (r *SubscriptionRepository) List(ctx context.Context) ([]domain.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+subscriptionColumns+` FROM calendar_subscriptions ORDER BY position, name`)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()
	var out []domain.Subscription
	for rows.Next() {
		s, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Get implements usecase.SubscriptionRepository.
func (r *SubscriptionRepository) Get(ctx context.Context, id string) (domain.Subscription, error) {
	s, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+subscriptionColumns+` FROM calendar_subscriptions WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Subscription{}, domain.ErrSubscriptionNotFound
	}
	return s, err
}

// Save implements usecase.SubscriptionRepository (insert or update).
func (r *SubscriptionRepository) Save(ctx context.Context, s domain.Subscription) error {
	enc, err := r.cipher.Encrypt(s.URL)
	if err != nil {
		return fmt.Errorf("encrypt subscription URL: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO calendar_subscriptions (id, name, url_encrypted, color, refresh_minutes, position)
		VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(position), -1) + 1 FROM calendar_subscriptions))
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, url_encrypted = excluded.url_encrypted, color = excluded.color,
			refresh_minutes = excluded.refresh_minutes`,
		s.ID, s.Name, enc, s.Color, int(s.RefreshInterval/time.Minute))
	if err != nil {
		return fmt.Errorf("upsert subscription: %w", err)
	}
	return nil
}

// Delete implements usecase.SubscriptionRepository.
func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM calendar_subscriptions WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	return nil
}
