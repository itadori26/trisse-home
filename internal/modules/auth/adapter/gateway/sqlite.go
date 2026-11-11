package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/sqlutil"
)

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AccountRepository stores accounts in SQLite. TOTP secrets are encrypted,
// recovery codes are stored as hashes only.
type AccountRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewAccountRepository returns a repository backed by db.
func NewAccountRepository(db *sql.DB, cipher Cipher) *AccountRepository {
	return &AccountRepository{db: db, cipher: cipher}
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

const accountColumns = `id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs`

// scan reads one account row without its recovery codes; callers load
// them with withRecoveryCodes after the row set is closed (SQLite with a
// single connection cannot run a second query while rows are open).
func (r *AccountRepository) scan(sc interface{ Scan(...any) error }) (domain.Account, error) {
	var id, role, name, hash, created, updated, secretEnc, pendingEnc, tabsJSON string
	var lastStep int64
	var disabled int
	if err := sc.Scan(&id, &role, &name, &hash, &created, &updated, &secretEnc, &pendingEnc, &lastStep, &disabled, &tabsJSON); err != nil {
		return domain.Account{}, err
	}
	createdAt, err := sqlutil.ParseTime(created)
	if err != nil {
		return domain.Account{}, err
	}
	updatedAt, err := sqlutil.ParseTime(updated)
	if err != nil {
		return domain.Account{}, err
	}
	var tabs []string
	if err := json.Unmarshal([]byte(tabsJSON), &tabs); err != nil {
		return domain.Account{}, fmt.Errorf("decode tabs: %w", err)
	}
	a := domain.RestoreAccount(id, domain.Role(role), name, hash, createdAt, updatedAt, disabled == 1, tabs)
	t := domain.TOTP{LastStep: lastStep}
	if t.Secret, err = r.decrypt(secretEnc); err != nil {
		return domain.Account{}, err
	}
	if t.PendingSecret, err = r.decrypt(pendingEnc); err != nil {
		return domain.Account{}, err
	}
	a.RestoreTwoFactor(t)
	return a, nil
}

func (r *AccountRepository) withRecoveryCodes(ctx context.Context, a domain.Account) (domain.Account, error) {
	codes, err := r.recoveryCodes(ctx, a.ID())
	if err != nil {
		return domain.Account{}, err
	}
	t := a.TwoFactor()
	t.RecoveryCodes = codes
	a.RestoreTwoFactor(t)
	return a, nil
}

const (
	selectAccountByID   = `SELECT ` + accountColumns + ` FROM auth_accounts WHERE id = ?`
	selectAccountByName = `SELECT ` + accountColumns + ` FROM auth_accounts WHERE name = ?`
	selectAccounts      = `SELECT ` + accountColumns + ` FROM auth_accounts ORDER BY role = 'guest', name`
)

func (r *AccountRepository) one(ctx context.Context, query string, arg any, notFound error) (domain.Account, error) {
	a, err := r.scan(r.db.QueryRowContext(ctx, query, arg))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Account{}, notFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("select account: %w", err)
	}
	return r.withRecoveryCodes(ctx, a)
}

// Get implements usecase.AccountRepository.
func (r *AccountRepository) Get(ctx context.Context, id string) (domain.Account, error) {
	notFound := domain.ErrGuestNotFound
	if id == domain.OwnerID {
		notFound = domain.ErrAccountNotFound
	}
	return r.one(ctx, selectAccountByID, id, notFound)
}

// FindByName implements usecase.AccountRepository.
func (r *AccountRepository) FindByName(ctx context.Context, name string) (domain.Account, error) {
	return r.one(ctx, selectAccountByName, name, domain.ErrGuestNotFound)
}

// List implements usecase.AccountRepository (owner first, then by name).
func (r *AccountRepository) List(ctx context.Context) ([]domain.Account, error) {
	out, err := r.listRows(ctx)
	if err != nil {
		return nil, err
	}
	// Recovery codes are loaded after the row set is closed.
	for i, a := range out {
		if out[i], err = r.withRecoveryCodes(ctx, a); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (r *AccountRepository) listRows(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.db.QueryContext(ctx, selectAccounts)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		a, err := r.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AccountRepository) recoveryCodes(ctx context.Context, accountID string) ([]domain.TokenHash, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code_hash FROM auth_recovery_codes WHERE account_id = ? ORDER BY code_hash`, accountID)
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

func tabsJSON(a domain.Account) string {
	b, _ := json.Marshal(a.Tabs()) // a slice of strings: cannot fail
	return string(b)
}

// Create implements usecase.AccountRepository.
func (r *AccountRepository) Create(ctx context.Context, a domain.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at, disabled, tabs) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID(), string(a.Role()), a.Name(), a.PasswordHash(), sqlutil.FormatTime(a.CreatedAt()), sqlutil.FormatTime(a.UpdatedAt()),
		sqlutil.BoolToInt(a.Disabled()), tabsJSON(a))
	if sqlutil.IsUniqueViolation(err) {
		if a.IsOwner() {
			return domain.ErrAlreadySetUp
		}
		return domain.ErrGuestNameTaken
	}
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	return r.Update(ctx, a)
}

// Update implements usecase.AccountRepository; the account row and its
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
		`UPDATE auth_accounts SET password_hash = ?, updated_at = ?, totp_secret_enc = ?, totp_pending_enc = ?, totp_last_step = ?,
		 disabled = ?, tabs = ? WHERE id = ?`,
		a.PasswordHash(), sqlutil.FormatTime(a.UpdatedAt()), secretEnc, pendingEnc, t.LastStep,
		sqlutil.BoolToInt(a.Disabled()), tabsJSON(a), a.ID())
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrAccountNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM auth_recovery_codes WHERE account_id = ?`, a.ID()); err != nil {
		return fmt.Errorf("clear recovery codes: %w", err)
	}
	for _, h := range t.RecoveryCodes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO auth_recovery_codes (account_id, code_hash) VALUES (?, ?)`, a.ID(), string(h)); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}
	return tx.Commit()
}

// Delete implements usecase.AccountRepository; the second factor goes with it.
func (r *AccountRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_recovery_codes WHERE account_id = ?`, id); err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_accounts WHERE id = ?`, id); err != nil {
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
		`INSERT INTO auth_sessions (account_id, token_hash, public_id, csrf_token, user_agent, client_ip, created_at, expires_at, refreshed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.AccountID(), string(s.TokenHash()), s.PublicID(), s.CSRFToken(), s.Client().UserAgent, s.Client().IP, sqlutil.FormatTime(s.CreatedAt()),
		sqlutil.FormatTime(s.ExpiresAt()), sqlutil.FormatTime(s.RefreshedAt()))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

const sessionColumns = `account_id, token_hash, public_id, csrf_token, user_agent, client_ip, created_at, expires_at, refreshed_at`

func scanSession(sc interface{ Scan(...any) error }) (domain.Session, error) {
	var account, hash, publicID, csrf, ua, ip, created, expires, refreshed string
	if err := sc.Scan(&account, &hash, &publicID, &csrf, &ua, &ip, &created, &expires, &refreshed); err != nil {
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
	return domain.RestoreSession(account, domain.TokenHash(hash), publicID, csrf, domain.Client{UserAgent: ua, IP: ip}, times[0], times[1], times[2]), nil
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
func (r *SessionRepository) ListActive(ctx context.Context, accountID string, now time.Time) ([]domain.Session, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+sessionColumns+` FROM auth_sessions WHERE account_id = ? AND expires_at > ? ORDER BY refreshed_at DESC`,
		accountID, sqlutil.FormatTime(now))
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
func (r *SessionRepository) DeleteByPublicID(ctx context.Context, accountID, publicID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE account_id = ? AND public_id = ?`, accountID, publicID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrSessionNotListed
	}
	return nil
}

// DeleteAllExcept implements usecase.SessionRepository.
func (r *SessionRepository) DeleteAllExcept(ctx context.Context, accountID string, keep domain.TokenHash) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE account_id = ? AND token_hash <> ?`, accountID, string(keep))
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

// DeleteAllFor implements usecase.SessionRepository.
func (r *SessionRepository) DeleteAllFor(ctx context.Context, accountID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
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
