package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"trisse-home/internal/modules/todos/domain"
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

const columns = `id, title, due, position, created_at, completed_at`

func scan(s interface{ Scan(...any) error }) (domain.Todo, error) {
	var id, title, created string
	var due, completed *string
	var pos int
	if err := s.Scan(&id, &title, &due, &pos, &created, &completed); err != nil {
		return domain.Todo{}, err
	}
	var times [3]time.Time
	var err error
	if times[0], err = sqlutil.ParseNullableTime(due); err != nil {
		return domain.Todo{}, err
	}
	if times[1], err = sqlutil.ParseTime(created); err != nil {
		return domain.Todo{}, err
	}
	if times[2], err = sqlutil.ParseNullableTime(completed); err != nil {
		return domain.Todo{}, err
	}
	return domain.RestoreTodo(id, title, times[0], pos, times[1], times[2]), nil
}

// List returns all to-dos.
func (r *Repository) List(ctx context.Context) ([]domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM todos ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()
	var out []domain.Todo
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Get returns one to-do.
func (r *Repository) Get(ctx context.Context, id string) (domain.Todo, error) {
	t, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM todos WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	if err != nil {
		return domain.Todo{}, fmt.Errorf("select todo: %w", err)
	}
	return t, nil
}

// Save inserts or updates a to-do.
func (r *Repository) Save(ctx context.Context, t domain.Todo) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO todos (`+columns+`) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, due = excluded.due,
			position = excluded.position, completed_at = excluded.completed_at`,
		t.ID(), t.Title(), sqlutil.FormatNullableTime(t.Due()), t.Position(),
		sqlutil.FormatTime(t.CreatedAt()), sqlutil.FormatNullableTime(t.CompletedAt()))
	if err != nil {
		return fmt.Errorf("upsert todo: %w", err)
	}
	return nil
}

// Delete removes the given to-dos.
func (r *Repository) Delete(ctx context.Context, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	if _, err := r.db.ExecContext(ctx, `DELETE FROM todos WHERE id IN (`+placeholders+`)`, args...); err != nil { //nolint:gosec // only placeholders are concatenated
		return fmt.Errorf("delete todos: %w", err)
	}
	return nil
}
