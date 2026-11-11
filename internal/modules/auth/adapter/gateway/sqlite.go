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

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AccountRepository stores the account in SQLite. TOTP secrets are
// encrypted, recovery codes are stored as hashes only.
type AccountRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewAccountRepository returns a repository backed by db.
func NewAccountRepository(db *sql.DB, cipher Cipher) *AccountRepository {
	return &AccountRepository{db: db, cipher: cipher}
}

// accountRow is the persistence model of domain.Account.
type accountRow struct {
	passwordHash, createdAt, updatedAt string
	secretEnc, pendingEnc              string
	lastStep                           int64
}

func (r *AccountRepository) decrypt(enc string) (string, error) {
	if enc == "" {
		return "", nil
	}
	plain, err := r.cipher.Decrypt(enc)
	if err != nil {
		return "", fmt.Errorf("decrypt totp secret (was SECRET_KEY changed?): %w", err)
	}
	return plain, nil
}

func (r *AccountRepository) encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return r.cipher.Encrypt(plain)
}

// Get implements usecase.AccountRepository.
func (r *AccountRepository) Get(ctx context.Context) (domain.Account, error) {
	var row accountRow
	err := r.db.QueryRowContext(ctx,
		`SELECT password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step FROM auth_account WHERE id = ?`,
		singleAccountID).Scan(&row.passwordHash, &row.createdAt, &row.updatedAt, &row.secretEnc, &row.pendingEnc, &row.lastStep)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("select account: %w", err)
	}
	created, err := sqlutil.ParseTime(row.createdAt)
	if err != nil {
		return domain.Account{}, err
	}
	updated, err := sqlutil.ParseTime(row.updatedAt)
	if err != nil {
		return domain.Account{}, err
	}
	a := domain.RestoreAccount(row.passwordHash, created, updated)
	t := domain.TOTP{LastStep: row.lastStep}
	if t.Secret, err = r.decrypt(row.secretEnc); err != nil {
		return domain.Account{}, err
	}
	if t.PendingSecret, err = r.decrypt(row.pendingEnc); err != nil {
		return domain.Account{}, err
	}
	if t.RecoveryCodes, err = r.recoveryCodes(ctx); err != nil {
		return domain.Account{}, err
	}
	a.RestoreTwoFactor(t)
	return a, nil
}

func (r *AccountRepository) recoveryCodes(ctx context.Context) ([]domain.TokenHash, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code_hash FROM auth_recovery_codes ORDER BY code_hash`)
	if err != nil {
		return nil, fmt.Errorf("query recovery codes: %w", err)
	}
	defer rows.Close()
	var out []domain.TokenHash
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, fmt.Errorf("scan recovery code: %w", err)
		}
		out = append(out, domain.TokenHash(h))
	}
	return out, rows.Err()
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
	return r.Update(ctx, a)
}

// Update implements usecase.AccountRepository; the account row and the
// recovery codes are written in one transaction.
func (r *AccountRepository) Update(ctx context.Context, a domain.Account) error {
	t := a.TwoFactor()
	secretEnc, err := r.encrypt(t.Secret)
	if err != nil {
		return fmt.Errorf("encrypt totp secret: %w", err)
	}
	pendingEnc, err := r.encrypt(t.PendingSecret)
	if err != nil {
		return fmt.Errorf("encrypt totp secret: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx,
		`UPDATE auth_account SET password_hash = ?, updated_at = ?, totp_secret_enc = ?, totp_pending_enc = ?, totp_last_step = ? WHERE id = ?`,
		a.PasswordHash(), sqlutil.FormatTime(a.UpdatedAt()), secretEnc, pendingEnc, t.LastStep, singleAccountID)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrAccountNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM auth_recovery_codes`); err != nil {
		return fmt.Errorf("clear recovery codes: %w", err)
	}
	for _, h := range t.RecoveryCodes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO auth_recovery_codes (code_hash) VALUES (?)`, string(h)); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}
	return tx.Commit()
}

// Delete implements usecase.AccountRepository; the second factor goes with it.
func (r *AccountRepository) Delete(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_recovery_codes`); err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}
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
		`INSERT INTO auth_sessions (token_hash, public_id, csrf_token, user_agent, client_ip, created_at, expires_at, refreshed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		string(s.TokenHash()), s.PublicID(), s.CSRFToken(), s.Client().UserAgent, s.Client().IP, sqlutil.FormatTime(s.CreatedAt()),
		sqlutil.FormatTime(s.ExpiresAt()), sqlutil.FormatTime(s.RefreshedAt()))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

const sessionColumns = `token_hash, public_id, csrf_token, user_agent, client_ip, created_at, expires_at, refreshed_at`

func scanSession(sc interface{ Scan(...any) error }) (domain.Session, error) {
	var hash, publicID, csrf, ua, ip, created, expires, refreshed string
	if err := sc.Scan(&hash, &publicID, &csrf, &ua, &ip, &created, &expires, &refreshed); err != nil {
		return domain.Session{}, err
	}
	times := make([]time.Time, 3)
	for i, raw := range []string{created, expires, refreshed} {
		t, err := sqlutil.ParseTime(raw)
		if err != nil {
			return domain.Session{}, err
		}
		times[i] = t
	}
	return domain.RestoreSession(domain.TokenHash(hash), publicID, csrf, domain.Client{UserAgent: ua, IP: ip}, times[0], times[1], times[2]), nil
}

// Find implements usecase.SessionRepository.
func (r *SessionRepository) Find(ctx context.Context, hash domain.TokenHash) (domain.Session, error) {
	s, err := scanSession(r.db.QueryRowContext(ctx, `SELECT `+sessionColumns+` FROM auth_sessions WHERE token_hash = ?`, string(hash)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("select session: %w", err)
	}
	return s, nil
}

// ListActive implements usecase.SessionRepository; most recently used first.
func (r *SessionRepository) ListActive(ctx context.Context, now time.Time) ([]domain.Session, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+sessionColumns+` FROM auth_sessions WHERE expires_at > ? ORDER BY refreshed_at DESC`, sqlutil.FormatTime(now))
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()
	var out []domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteByPublicID implements usecase.SessionRepository.
func (r *SessionRepository) DeleteByPublicID(ctx context.Context, publicID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE public_id = ?`, publicID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrSessionNotListed
	}
	return nil
}

// DeleteAllExcept implements usecase.SessionRepository.
func (r *SessionRepository) DeleteAllExcept(ctx context.Context, keep domain.TokenHash) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash <> ?`, string(keep))
	if err != nil {
		return 0, fmt.Errorf("delete sessions: %w", err)
	}
	return res.RowsAffected()
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
