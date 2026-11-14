// Package gateway stores drop items and their files in SQLite.
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

	"trisse-home/internal/modules/drop/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Repository. Items are short-lived and not
// part of data exports.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, kind, text, file_name, content_type, size, created_at, expires_at`

func scan(row interface{ Scan(...any) error }) (domain.Item, error) {
	var i domain.Item
	var kind, created, expires string
	if err := row.Scan(&i.ID, &kind, &i.Text, &i.FileName, &i.ContentType, &i.Size, &created, &expires); err != nil {
		return domain.Item{}, err
	}
	i.Kind = domain.Kind(kind)
	var err error
	if i.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
		return domain.Item{}, err
	}
	i.ExpiresAt, err = sqlutil.ParseTime(expires)
	return i, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Item, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM drop_items`)
	if err != nil {
		return nil, fmt.Errorf("select items: %w", err)
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		i, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Item, error) {
	i, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM drop_items WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("select item: %w", err)
	}
	return i, nil
}

// Add implements usecase.Repository.
func (r *Repository) Add(ctx context.Context, i domain.Item, data []byte) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `INSERT INTO drop_items (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		i.ID, string(i.Kind), i.Text, i.FileName, i.ContentType, i.Size, sqlutil.FormatTime(i.CreatedAt), sqlutil.FormatTime(i.ExpiresAt)); err != nil {
		return fmt.Errorf("insert item: %w", err)
	}
	if data != nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO drop_files (item_id, data) VALUES (?, ?)`, i.ID, data); err != nil {
			return fmt.Errorf("insert file: %w", err)
		}
	}
	return tx.Commit()
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM drop_files WHERE item_id = ?`, id); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM drop_items WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}

// File implements usecase.Repository.
func (r *Repository) File(ctx context.Context, id string) ([]byte, error) {
	var data []byte
	err := r.db.QueryRowContext(ctx, `SELECT data FROM drop_files WHERE item_id = ?`, id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select file: %w", err)
	}
	return data, nil
}

// DeleteExpired implements usecase.Repository.
func (r *Repository) DeleteExpired(ctx context.Context, now time.Time) (int, error) {
	at := sqlutil.FormatTime(now)
	if _, err := r.db.ExecContext(ctx, `DELETE FROM drop_files WHERE item_id IN (SELECT id FROM drop_items WHERE expires_at <= ?)`, at); err != nil {
		return 0, fmt.Errorf("delete expired files: %w", err)
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM drop_items WHERE expires_at <= ?`, at)
	if err != nil {
		return 0, fmt.Errorf("delete expired items: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
