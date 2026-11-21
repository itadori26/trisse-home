// Package gateway stores chores in SQLite.
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
	"strings"
	"time"

	"waylight/internal/modules/chores/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{{Name: "chores_items"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, title, every_days, due, assignees, turn, last_done_at, last_done_by, owner_id, visibility, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Chore, error) {
	var c domain.Chore
	var due, assignees, lastDone, owner, vis, created string
	if err := sc.Scan(&c.ID, &c.Title, &c.EveryDays, &due, &assignees, &c.Turn, &lastDone, &c.LastDoneBy, &owner, &vis, &created); err != nil {
		return c, err
	}
	var err error
	if c.Due, err = time.Parse(time.DateOnly, due); err != nil {
		return c, err
	}
	if assignees != "" {
		c.Assignees = strings.Split(assignees, "\n")
	}
	if lastDone != "" {
		if c.LastDoneAt, err = sqlutil.ParseTime(lastDone); err != nil {
			return c, err
		}
	}
	c.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	c.CreatedAt, err = sqlutil.ParseTime(created)
	return c, err
}

func formatOptional(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return sqlutil.FormatTime(t)
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Chore, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM chores_items WHERE `+cond+` ORDER BY due, title`, args...)
	if err != nil {
		return nil, fmt.Errorf("query chores: %w", err)
	}
	defer rows.Close()
	var out []domain.Chore
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan chore: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Chore, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	c, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM chores_items WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Chore{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Chore{}, fmt.Errorf("select chore: %w", err)
	}
	return c, nil
}

// Save implements usecase.Repository (upsert). A new chore belongs to the
// principal; an existing one is changed only by its owner (ADR 0018).
func (r *Repository) Save(ctx context.Context, c domain.Chore) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "chores_items")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO chores_items (id, title, every_days, due, assignees, turn, last_done_at, last_done_by, owner_id, visibility, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, every_days = excluded.every_days, due = excluded.due,
		assignees = excluded.assignees, turn = excluded.turn, visibility = excluded.visibility
		WHERE `+owned,
		append([]any{c.ID, c.Title, c.EveryDays, c.DueDate(), strings.Join(c.Assignees, "\n"), c.Turn,
			formatOptional(c.LastDoneAt), c.LastDoneBy, p.AccountID, string(c.Access.Visibility), sqlutil.FormatTime(c.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert chore: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// SaveDone implements usecase.Repository: it writes only the fields a
// completion changes, for anyone who sees the chore, and only while the
// chore is still due on prevDue.
func (r *Repository) SaveDone(ctx context.Context, c domain.Chore, prevDue time.Time) (bool, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `UPDATE chores_items SET due = ?, turn = ?, last_done_at = ?, last_done_by = ?
		WHERE id = ? AND due = ? AND `+cond,
		append([]any{c.DueDate(), c.Turn, formatOptional(c.LastDoneAt), c.LastDoneBy, c.ID, prevDue.Format(time.DateOnly)}, args...)...)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// Delete implements usecase.Repository for the owner; a visible chore of
// someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM chores_items WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete chore: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}
