package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// KioskRepository implements usecase.KioskRepository (#50).
type KioskRepository struct{ db *sql.DB }

// NewKioskRepository returns a repository backed by db.
func NewKioskRepository(db *sql.DB) *KioskRepository { return &KioskRepository{db: db} }

const kioskColumns = `account_id, token_hash, rotate_seconds, night_from, night_to, last_used_at, last_used_ip, loopback_only`

func scanKiosk(row interface{ Scan(...any) error }) (domain.Kiosk, error) {
	var k domain.Kiosk
	var hash string
	var lastUsed *string
	err := row.Scan(&k.AccountID, &hash, &k.Options.RotateSeconds, &k.Options.NightFrom, &k.Options.NightTo, &lastUsed, &k.LastUsedIP, &k.LoopbackOnly)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Kiosk{}, domain.ErrKioskNotFound
	}
	if err != nil {
		return domain.Kiosk{}, fmt.Errorf("select kiosk: %w", err)
	}
	k.TokenHash = domain.TokenHash(hash)
	k.LastUsedAt, err = sqlutil.ParseNullableTime(lastUsed)
	return k, err
}

// Save implements usecase.KioskRepository (upsert).
func (r *KioskRepository) Save(ctx context.Context, k domain.Kiosk) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO auth_kiosks (`+kioskColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (account_id) DO UPDATE SET token_hash = excluded.token_hash, rotate_seconds = excluded.rotate_seconds,
			night_from = excluded.night_from, night_to = excluded.night_to, last_used_at = excluded.last_used_at, last_used_ip = excluded.last_used_ip,
			loopback_only = excluded.loopback_only`,
		k.AccountID, string(k.TokenHash), k.Options.RotateSeconds, k.Options.NightFrom, k.Options.NightTo, sqlutil.FormatNullableTime(k.LastUsedAt), k.LastUsedIP,
		k.LoopbackOnly)
	if err != nil {
		return fmt.Errorf("upsert kiosk: %w", err)
	}
	return nil
}

// Get implements usecase.KioskRepository.
func (r *KioskRepository) Get(ctx context.Context, accountID string) (domain.Kiosk, error) {
	return scanKiosk(r.db.QueryRowContext(ctx, `SELECT `+kioskColumns+` FROM auth_kiosks WHERE account_id = ?`, accountID))
}

// FindByTokenHash implements usecase.KioskRepository.
func (r *KioskRepository) FindByTokenHash(ctx context.Context, hash domain.TokenHash) (domain.Kiosk, error) {
	return scanKiosk(r.db.QueryRowContext(ctx, `SELECT `+kioskColumns+` FROM auth_kiosks WHERE token_hash = ?`, string(hash)))
}

// Delete implements usecase.KioskRepository.
func (r *KioskRepository) Delete(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_kiosks WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("delete kiosk: %w", err)
	}
	return nil
}
