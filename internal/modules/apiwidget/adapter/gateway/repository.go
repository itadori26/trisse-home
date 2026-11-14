// Package gateway stores the endpoints in SQLite, requests them over HTTP
// and evaluates JSONPath.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are exported without the encrypted header value, which
// only works with this instance's SECRET_KEY.
var TransferTables = []sqlutil.Table{{Name: "apiwidget_endpoints", Exclude: []string{"header_value_encrypted"}}}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// Repository implements usecase.EndpointRepository.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

type extractionRow struct {
	Label string   `json:"label"`
	Path  string   `json:"path"`
	Unit  string   `json:"unit"`
	Warn  *float64 `json:"warn,omitempty"`
	Crit  *float64 `json:"crit,omitempty"`
	Below bool     `json:"below,omitempty"`
}

const selectEndpoints = `SELECT id, name, url, header_name, header_value_encrypted, interval_seconds, allow_public, extractions FROM apiwidget_endpoints`

// List implements usecase.EndpointRepository, in creation order.
func (r *Repository) List(ctx context.Context) ([]domain.Endpoint, error) {
	rows, err := r.db.QueryContext(ctx, selectEndpoints+` ORDER BY position, rowid`)
	if err != nil {
		return nil, fmt.Errorf("select endpoints: %w", err)
	}
	defer rows.Close()
	out := []domain.Endpoint{}
	for rows.Next() {
		e, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Get implements usecase.EndpointRepository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Endpoint, error) {
	e, err := r.scan(r.db.QueryRowContext(ctx, selectEndpoints+` WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Endpoint{}, domain.ErrNotFound
	}
	return e, err
}

func (r *Repository) scan(row interface{ Scan(...any) error }) (domain.Endpoint, error) {
	var e domain.Endpoint
	var enc, raw string
	var seconds int64
	if err := row.Scan(&e.ID, &e.Name, &e.URL, &e.HeaderName, &enc, &seconds, &e.AllowPublic, &raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Endpoint{}, err
		}
		return domain.Endpoint{}, fmt.Errorf("scan endpoint: %w", err)
	}
	e.Interval = time.Duration(seconds) * time.Second
	if enc != "" {
		v, err := r.cipher.Decrypt(enc)
		if err != nil {
			return domain.Endpoint{}, fmt.Errorf("decrypt header value (was SECRET_KEY changed?): %w", err)
		}
		e.HeaderValue = v
	} else if e.HeaderName != "" {
		e.HeaderName = "" // imported without its secret: the header has to be entered again
	}
	var xs []extractionRow
	if err := json.Unmarshal([]byte(raw), &xs); err != nil {
		return domain.Endpoint{}, fmt.Errorf("decode extractions: %w", err)
	}
	for _, x := range xs {
		e.Extractions = append(e.Extractions, domain.Extraction(x))
	}
	return e, nil
}

// Save implements usecase.EndpointRepository (insert or update).
func (r *Repository) Save(ctx context.Context, e domain.Endpoint) error {
	enc := ""
	if e.HeaderValue != "" {
		var err error
		if enc, err = r.cipher.Encrypt(e.HeaderValue); err != nil {
			return fmt.Errorf("encrypt header value: %w", err)
		}
	}
	xs := make([]extractionRow, 0, len(e.Extractions))
	for _, x := range e.Extractions {
		xs = append(xs, extractionRow(x))
	}
	raw, err := json.Marshal(xs)
	if err != nil {
		return fmt.Errorf("encode extractions: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO apiwidget_endpoints
		(id, name, url, header_name, header_value_encrypted, interval_seconds, allow_public, extractions, position)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(position), 0) + 1 FROM apiwidget_endpoints))
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, url = excluded.url, header_name = excluded.header_name,
			header_value_encrypted = excluded.header_value_encrypted, interval_seconds = excluded.interval_seconds,
			allow_public = excluded.allow_public, extractions = excluded.extractions`,
		e.ID, e.Name, e.URL, e.HeaderName, enc, int64(e.Interval/time.Second), e.AllowPublic, string(raw))
	if err != nil {
		return fmt.Errorf("upsert endpoint: %w", err)
	}
	return nil
}

// Delete implements usecase.EndpointRepository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM apiwidget_endpoints WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete endpoint: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
