package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// TokenRepository implements usecase.TokenRepository in SQLite.
type TokenRepository struct{ db *sql.DB }

// NewTokenRepository returns the API token store.
func NewTokenRepository(db *sql.DB) *TokenRepository { return &TokenRepository{db: db} }

func optionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return sqlutil.FormatTime(t)
}

func parseOptional(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return sqlutil.ParseTime(s)
}

const tokenColumns = `id, name, token_hash, scopes, created_at, expires_at, last_used_at, last_used_ip` //nolint:gosec // column names, not a credential

func scanToken(sc interface{ Scan(...any) error }) (domain.APIToken, error) {
	var t domain.APIToken
	var hash, scopes, created, expires, used string
	if err := sc.Scan(&t.ID, &t.Name, &hash, &scopes, &created, &expires, &used, &t.LastUsedIP); err != nil {
		return t, err
	}
	t.Hash, t.Scopes = domain.TokenHash(hash), strings.Split(scopes, " ")
	var err error
	if t.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
		return t, err
	}
	if t.ExpiresAt, err = parseOptional(expires); err != nil {
		return t, err
	}
	t.LastUsedAt, err = parseOptional(used)
	return t, err
}

// Create implements usecase.TokenRepository.
func (r *TokenRepository) Create(ctx context.Context, t domain.APIToken) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO auth_api_tokens (`+tokenColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, string(t.Hash), strings.Join(t.Scopes, " "), sqlutil.FormatTime(t.CreatedAt), optionalTime(t.ExpiresAt),
		optionalTime(t.LastUsedAt), t.LastUsedIP)
	if err != nil {
		return fmt.Errorf("insert token: %w", err)
	}
	return nil
}

// List implements usecase.TokenRepository (newest first).
func (r *TokenRepository) List(ctx context.Context) ([]domain.APIToken, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+tokenColumns+` FROM auth_api_tokens ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query tokens: %w", err)
	}
	defer rows.Close()
	var out []domain.APIToken
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, fmt.Errorf("scan token: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// FindByHash implements usecase.TokenRepository.
func (r *TokenRepository) FindByHash(ctx context.Context, h domain.TokenHash) (domain.APIToken, error) {
	t, err := scanToken(r.db.QueryRowContext(ctx, `SELECT `+tokenColumns+` FROM auth_api_tokens WHERE token_hash = ?`, string(h)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.APIToken{}, domain.ErrSessionInvalid
	}
	if err != nil {
		return domain.APIToken{}, fmt.Errorf("select token: %w", err)
	}
	return t, nil
}

// UpdateUse implements usecase.TokenRepository.
func (r *TokenRepository) UpdateUse(ctx context.Context, t domain.APIToken) error {
	_, err := r.db.ExecContext(ctx, `UPDATE auth_api_tokens SET last_used_at = ?, last_used_ip = ? WHERE id = ?`,
		optionalTime(t.LastUsedAt), t.LastUsedIP, t.ID)
	if err != nil {
		return fmt.Errorf("update token: %w", err)
	}
	return nil
}

// Delete implements usecase.TokenRepository.
func (r *TokenRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_api_tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}
