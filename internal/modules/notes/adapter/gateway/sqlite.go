// Package gateway implements the notes repository on SQLite.
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

	"trisse-home/internal/modules/notes/domain"
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

const columns = `id, title, body, pinned, created_at, updated_at`

func scan(s interface{ Scan(...any) error }) (domain.Note, error) {
	var id, title, body, created, updated string
	var pinned int
	if err := s.Scan(&id, &title, &body, &pinned, &created, &updated); err != nil {
		return domain.Note{}, err
	}
	c, err := sqlutil.ParseTime(created)
	if err != nil {
		return domain.Note{}, err
	}
	u, err := sqlutil.ParseTime(updated)
	if err != nil {
		return domain.Note{}, err
	}
	return domain.RestoreNote(id, title, body, pinned == 1, c, u), nil
}

// List returns all notes, pinned first, then by last edit.
func (r *Repository) List(ctx context.Context) ([]domain.Note, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM notes ORDER BY pinned DESC, updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()
	var out []domain.Note
	for rows.Next() {
		n, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// Get returns one note.
func (r *Repository) Get(ctx context.Context, id string) (domain.Note, error) {
	n, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM notes WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Note{}, domain.ErrNoteNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("select note: %w", err)
	}
	return n, nil
}

// Save inserts or updates a note.
func (r *Repository) Save(ctx context.Context, n domain.Note) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO notes (`+columns+`) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, body = excluded.body,
			pinned = excluded.pinned, updated_at = excluded.updated_at`,
		n.ID(), n.Title(), n.Body(), sqlutil.BoolToInt(n.Pinned()), sqlutil.FormatTime(n.CreatedAt()), sqlutil.FormatTime(n.UpdatedAt()))
	if err != nil {
		return fmt.Errorf("upsert note: %w", err)
	}
	return nil
}

// Delete removes a note.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}
