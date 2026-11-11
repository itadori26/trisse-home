package demo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// ErrSecretsDisabled is returned when something tries to store a secret.
var ErrSecretsDisabled = kernel.NewPublicError(kernel.ErrForbidden, "Storing credentials is disabled in the demo.")

// Cipher is the secret store interface the modules use.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// ReadOnlySecrets refuses to encrypt (and thus store) new secrets.
type ReadOnlySecrets struct{ Inner Cipher }

// Encrypt implements Cipher and always fails.
func (ReadOnlySecrets) Encrypt(string) (string, error) { return "", ErrSecretsDisabled }

// Decrypt implements Cipher.
func (s ReadOnlySecrets) Decrypt(c string) (string, error) { return s.Inner.Decrypt(c) }

const markerTable = "demo_instance"

// CheckDatabase refuses a database that already holds an account which was
// not created by demo mode, i.e. a production database.
func CheckDatabase(ctx context.Context, db *sql.DB) error {
	exists := func(table string) (bool, error) {
		var n int
		err := db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&n)
		return n > 0, err
	}
	hasAccounts, err := exists("auth_account")
	if err != nil {
		return fmt.Errorf("inspect database: %w", err)
	}
	if !hasAccounts {
		return nil
	}
	var accounts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM auth_account`).Scan(&accounts); err != nil {
		return fmt.Errorf("inspect database: %w", err)
	}
	marked, err := exists(markerTable)
	if err != nil {
		return fmt.Errorf("inspect database: %w", err)
	}
	if accounts > 0 && !marked {
		return errors.New("DEMO_MODE refuses to start on a database that already contains an account; use an empty data directory")
	}
	return nil
}

// MarkDatabase records that the database belongs to a demo instance.
func MarkDatabase(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS `+markerTable+` (id INTEGER PRIMARY KEY)`)
	return err
}

// blockedRoutes change the account or write files; visitors must not.
var blockedRoutes = []struct{ method, prefix string }{
	{http.MethodPut, "/api/auth/password"},
	{http.MethodPost, "/api/auth/totp/"},
	{http.MethodPost, "/api/auth/sessions/"},
	{http.MethodDelete, "/api/auth/sessions/"},
	{http.MethodPost, "/api/auth/logout"},
	{http.MethodPost, "/api/backups"},
	{http.MethodPost, "/api/transfer/import"},
	{http.MethodPut, "/api/system/update-check"},
}

// Block answers blocked routes with 403 Forbidden.
func Block(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, b := range blockedRoutes {
			if r.Method == b.method && strings.HasPrefix(r.URL.Path, b.prefix) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"title":"Forbidden","status":403,"detail":"This action is disabled in the demo."}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// NoWake refuses Wake-on-LAN: a demo must not send packets into a network.
type NoWake struct{}

// Wake implements the services Waker port.
func (NoWake) Wake(context.Context, kernel.MAC) error {
	return kernel.NewPublicError(kernel.ErrForbidden, "Wake-on-LAN is disabled in the demo.")
}
