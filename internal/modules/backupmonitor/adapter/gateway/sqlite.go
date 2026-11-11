// Package gateway stores backup sources in SQLite.
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
	"time"

	"trisse-home/internal/modules/backupmonitor/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, name, max_age_seconds, token_hash, created_at, last_success_at, last_failure_at, last_message`

func scan(sc interface{ Scan(...any) error }) (domain.Source, error) {
	var s domain.Source
	var maxAge int64
	var token, created string
	var success, failure *string
	if err := sc.Scan(&s.ID, &s.Name, &maxAge, &token, &created, &success, &failure, &s.LastMessage); err != nil {
		return s, err
	}
	s.MaxAge, s.TokenHash = time.Duration(maxAge)*time.Second, domain.TokenHash(token)
	var err error
	if s.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
		return s, err
	}
	if s.LastSuccessAt, err = sqlutil.ParseNullableTime(success); err != nil {
		return s, err
	}
	s.LastFailureAt, err = sqlutil.ParseNullableTime(failure)
	return s, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Source, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM backup_sources ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query sources: %w", err)
	}
	defer rows.Close()
	var out []domain.Source
	for rows.Next() {
		s, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) one(ctx context.Context, where string, arg any) (domain.Source, error) {
	s, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM backup_sources WHERE `+where+` = ?`, arg))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Source{}, domain.ErrSourceNotFound
	}
	if err != nil {
		return domain.Source{}, fmt.Errorf("select source: %w", err)
	}
	return s, nil
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Source, error) {
	return r.one(ctx, "id", id)
}

// GetByToken implements usecase.Repository.
func (r *Repository) GetByToken(ctx context.Context, h domain.TokenHash) (domain.Source, error) {
	return r.one(ctx, "token_hash", string(h))
}

// Save implements usecase.Repository.
func (r *Repository) Save(ctx context.Context, s domain.Source) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO backup_sources (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, max_age_seconds = excluded.max_age_seconds,
			token_hash = excluded.token_hash, last_success_at = excluded.last_success_at,
			last_failure_at = excluded.last_failure_at, last_message = excluded.last_message`,
		s.ID, s.Name, int64(s.MaxAge/time.Second), string(s.TokenHash), sqlutil.FormatTime(s.CreatedAt),
		sqlutil.FormatNullableTime(s.LastSuccessAt), sqlutil.FormatNullableTime(s.LastFailureAt), s.LastMessage)
	if err != nil {
		return fmt.Errorf("upsert source: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM backup_sources WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}
	return nil
}
