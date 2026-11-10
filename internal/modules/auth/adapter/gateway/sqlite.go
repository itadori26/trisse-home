package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// singleAccountID is the fixed primary key of the only account row.
const singleAccountID = 1

// AccountRepository stores the account in SQLite.
type AccountRepository struct{ db *sql.DB }

// NewAccountRepository returns a repository backed by db.
func NewAccountRepository(db *sql.DB) *AccountRepository { return &AccountRepository{db: db} }

// accountRow is the persistence model of domain.Account.
type accountRow struct {
	passwordHash, createdAt, updatedAt string
}

func (r accountRow) toDomain() (domain.Account, error) {
	created, err := sqlutil.ParseTime(r.createdAt)
	if err != nil {
		return domain.Account{}, err
	}
	updated, err := sqlutil.ParseTime(r.updatedAt)
	if err != nil {
		return domain.Account{}, err
	}
	return domain.RestoreAccount(r.passwordHash, created, updated), nil
}

// Get implements usecase.AccountRepository.
func (r *AccountRepository) Get(ctx context.Context) (domain.Account, error) {
	var row accountRow
	err := r.db.QueryRowContext(ctx,
		`SELECT password_hash, created_at, updated_at FROM auth_account WHERE id = ?`, singleAccountID).
		Scan(&row.passwordHash, &row.createdAt, &row.updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("select account: %w", err)
	}
	return row.toDomain()
}

// Create implements usecase.AccountRepository.
func (r *AccountRepository) Create(ctx context.Context, a domain.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO auth_account (id, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		singleAccountID, a.PasswordHash(), sqlutil.FormatTime(a.CreatedAt()), sqlutil.FormatTime(a.UpdatedAt()))
	if sqlutil.IsUniqueViolation(err) {
		return domain.ErrAlreadySetUp
	}
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

// Update implements usecase.AccountRepository.
func (r *AccountRepository) Update(ctx context.Context, a domain.Account) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE auth_account SET password_hash = ?, updated_at = ? WHERE id = ?`,
		a.PasswordHash(), sqlutil.FormatTime(a.UpdatedAt()), singleAccountID)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

// Delete implements usecase.AccountRepository.
func (r *AccountRepository) Delete(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_account`); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// SessionRepository stores sessions in SQLite.
type SessionRepository struct{ db *sql.DB }

// NewSessionRepository returns a repository backed by db.
func NewSessionRepository(db *sql.DB) *SessionRepository { return &SessionRepository{db: db} }

// Create implements usecase.SessionRepository.
func (r *SessionRepository) Create(ctx context.Context, s domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO auth_sessions (token_hash, csrf_token, created_at, expires_at, refreshed_at) VALUES (?, ?, ?, ?, ?)`,
		string(s.TokenHash()), s.CSRFToken(), sqlutil.FormatTime(s.CreatedAt()),
		sqlutil.FormatTime(s.ExpiresAt()), sqlutil.FormatTime(s.RefreshedAt()))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Find implements usecase.SessionRepository.
func (r *SessionRepository) Find(ctx context.Context, hash domain.TokenHash) (domain.Session, error) {
	var csrf, created, expires, refreshed string
	err := r.db.QueryRowContext(ctx,
		`SELECT csrf_token, created_at, expires_at, refreshed_at FROM auth_sessions WHERE token_hash = ?`, string(hash)).
		Scan(&csrf, &created, &expires, &refreshed)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("select session: %w", err)
	}
	times := make([]time.Time, 3)
	for i, raw := range []string{created, expires, refreshed} {
		if times[i], err = sqlutil.ParseTime(raw); err != nil {
			return domain.Session{}, err
		}
	}
	return domain.RestoreSession(hash, csrf, times[0], times[1], times[2]), nil
}

// Update implements usecase.SessionRepository.
func (r *SessionRepository) Update(ctx context.Context, s domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE auth_sessions SET expires_at = ?, refreshed_at = ? WHERE token_hash = ?`,
		sqlutil.FormatTime(s.ExpiresAt()), sqlutil.FormatTime(s.RefreshedAt()), string(s.TokenHash()))
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

// Delete implements usecase.SessionRepository.
func (r *SessionRepository) Delete(ctx context.Context, hash domain.TokenHash) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, string(hash)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteAll implements usecase.SessionRepository.
func (r *SessionRepository) DeleteAll(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions`); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}
	return nil
}

// DeleteExpired implements usecase.SessionRepository.
func (r *SessionRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE expires_at <= ?`, sqlutil.FormatTime(now))
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return res.RowsAffected()
}
