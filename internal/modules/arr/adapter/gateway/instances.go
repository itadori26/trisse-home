// Package gateway stores *arr instances in SQLite and talks to Sonarr,
// Radarr and Prowlarr.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/arr/domain"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// InstanceRepository stores instances with encrypted API keys.
type InstanceRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewInstanceRepository returns a repository backed by db.
func NewInstanceRepository(db *sql.DB, cipher Cipher) *InstanceRepository {
	return &InstanceRepository{db: db, cipher: cipher}
}

const columns = `id, app, name, url, api_key_encrypted`

func (r *InstanceRepository) scan(row interface{ Scan(...any) error }) (domain.Instance, error) {
	var in domain.Instance
	var app, enc string
	if err := row.Scan(&in.ID, &app, &in.Name, &in.URL, &enc); err != nil {
		return domain.Instance{}, err
	}
	key, err := r.cipher.Decrypt(enc)
	if err != nil {
		return domain.Instance{}, fmt.Errorf("decrypt API key (was SECRET_KEY changed?): %w", err)
	}
	in.App, in.APIKey = domain.App(app), key
	return in, nil
}

// List implements usecase.InstanceRepository in creation order.
func (r *InstanceRepository) List(ctx context.Context) ([]domain.Instance, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM arr_instances ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("query instances: %w", err)
	}
	defer rows.Close()
	var out []domain.Instance
	for rows.Next() {
		in, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, in)
	}
	return out, rows.Err()
}

// Get implements usecase.InstanceRepository.
func (r *InstanceRepository) Get(ctx context.Context, id string) (domain.Instance, error) {
	in, err := r.scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM arr_instances WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Instance{}, domain.ErrInstanceNotFound
	}
	return in, err
}

// Save implements usecase.InstanceRepository (insert or update).
func (r *InstanceRepository) Save(ctx context.Context, in domain.Instance) error {
	enc, err := r.cipher.Encrypt(in.APIKey)
	if err != nil {
		return fmt.Errorf("encrypt API key: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO arr_instances (id, app, name, url, api_key_encrypted, position)
		VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(position), -1) + 1 FROM arr_instances))
		ON CONFLICT (id) DO UPDATE SET app = excluded.app, name = excluded.name, url = excluded.url,
			api_key_encrypted = excluded.api_key_encrypted`,
		in.ID, string(in.App), in.Name, in.URL, enc)
	if err != nil {
		return fmt.Errorf("upsert instance: %w", err)
	}
	return nil
}

// Delete implements usecase.InstanceRepository.
func (r *InstanceRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM arr_instances WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}
	return nil
}
