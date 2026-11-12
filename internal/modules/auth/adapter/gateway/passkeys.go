package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
)

// PasskeyRepository implements usecase.PasskeyRepository in SQLite.
type PasskeyRepository struct{ db *sql.DB }

// NewPasskeyRepository returns the passkey store.
func NewPasskeyRepository(db *sql.DB) *PasskeyRepository { return &PasskeyRepository{db: db} }

const passkeyColumns = `id, account_id, credential_id, public_key, alg, sign_count, name, created_at, last_used_at` //nolint:gosec // column names, not a credential

func scanPasskey(sc interface{ Scan(...any) error }) (domain.Passkey, error) {
	var p domain.Passkey
	var created, used string
	var count int64
	if err := sc.Scan(&p.ID, &p.AccountID, &p.CredentialID, &p.PublicKey, &p.Alg, &count, &p.Name, &created, &used); err != nil {
		return p, err
	}
	p.SignCount = uint32(count) //nolint:gosec // stored from a uint32
	var err error
	if p.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
		return p, err
	}
	p.LastUsedAt, err = parseOptional(used)
	return p, err
}

// Create implements usecase.PasskeyRepository.
func (r *PasskeyRepository) Create(ctx context.Context, p domain.Passkey) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO auth_passkeys (`+passkeyColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.AccountID, p.CredentialID, p.PublicKey, p.Alg, int64(p.SignCount), p.Name, sqlutil.FormatTime(p.CreatedAt), optionalTime(p.LastUsedAt))
	if sqlutil.IsUniqueViolation(err) {
		return kernel.NewPublicError(kernel.ErrConflict, "This passkey is already registered.")
	}
	if err != nil {
		return fmt.Errorf("insert passkey: %w", err)
	}
	return nil
}

// ListByAccount implements usecase.PasskeyRepository.
func (r *PasskeyRepository) ListByAccount(ctx context.Context, accountID string) ([]domain.Passkey, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+passkeyColumns+` FROM auth_passkeys WHERE account_id = ? ORDER BY created_at`, accountID)
	if err != nil {
		return nil, fmt.Errorf("query passkeys: %w", err)
	}
	defer rows.Close()
	var out []domain.Passkey
	for rows.Next() {
		p, err := scanPasskey(rows)
		if err != nil {
			return nil, fmt.Errorf("scan passkey: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// FindByCredentialID implements usecase.PasskeyRepository.
func (r *PasskeyRepository) FindByCredentialID(ctx context.Context, id []byte) (domain.Passkey, error) {
	p, err := scanPasskey(r.db.QueryRowContext(ctx, `SELECT `+passkeyColumns+` FROM auth_passkeys WHERE credential_id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Passkey{}, domain.ErrPasskeyNotFound
	}
	if err != nil {
		return domain.Passkey{}, fmt.Errorf("select passkey: %w", err)
	}
	return p, nil
}

// Update implements usecase.PasskeyRepository (counter and last use).
func (r *PasskeyRepository) Update(ctx context.Context, p domain.Passkey) error {
	_, err := r.db.ExecContext(ctx, `UPDATE auth_passkeys SET sign_count = ?, last_used_at = ? WHERE id = ?`,
		int64(p.SignCount), optionalTime(p.LastUsedAt), p.ID)
	if err != nil {
		return fmt.Errorf("update passkey: %w", err)
	}
	return nil
}

// Delete implements usecase.PasskeyRepository.
func (r *PasskeyRepository) Delete(ctx context.Context, accountID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_passkeys WHERE account_id = ? AND id = ?`, accountID, id)
	if err != nil {
		return fmt.Errorf("delete passkey: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrPasskeyNotFound
	}
	return nil
}

// DeleteByAccount removes all passkeys of an account (guest deleted).
func (r *PasskeyRepository) DeleteByAccount(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_passkeys WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("delete passkeys: %w", err)
	}
	return nil
}
