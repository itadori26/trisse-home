// Package gateway stores gift ideas in SQLite.
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

	"waylight/internal/modules/gifts/domain"
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
var TransferTables = []sqlutil.Table{{Name: "gifts_ideas"}}

// Repository implements usecase.Repository. Every statement, reads
// included, is limited with sqlutil.OwnedBy: ideas are never shared, not
// even with everyone (ADR 0018).
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, person, title, link, note, bought, owner_id, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Idea, error) {
	var i domain.Idea
	var bought int
	var created string
	if err := sc.Scan(&i.ID, &i.Person, &i.Title, &i.Link, &i.Note, &bought, &i.Owner, &created); err != nil {
		return i, err
	}
	i.Bought = bought == 1
	var err error
	i.CreatedAt, err = sqlutil.ParseTime(created)
	return i, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Idea, error) {
	cond, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM gifts_ideas WHERE `+cond+` ORDER BY person, created_at`, args...)
	if err != nil {
		return nil, fmt.Errorf("query gift ideas: %w", err)
	}
	defer rows.Close()
	var out []domain.Idea
	for rows.Next() {
		i, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan gift idea: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Idea, error) {
	cond, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	i, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM gifts_ideas WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Idea{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Idea{}, fmt.Errorf("select gift idea: %w", err)
	}
	return i, nil
}

// Save implements usecase.Repository (upsert). A new idea belongs to the
// principal; someone else's idea is reported as not found.
func (r *Repository) Save(ctx context.Context, i domain.Idea) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "gifts_ideas")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO gifts_ideas (id, person, title, link, note, bought, owner_id, visibility, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'private', ?)
		ON CONFLICT (id) DO UPDATE SET person = excluded.person, title = excluded.title, link = excluded.link,
		note = excluded.note, bought = excluded.bought
		WHERE `+owned,
		append([]any{i.ID, i.Person, i.Title, i.Link, i.Note, sqlutil.BoolToInt(i.Bought), p.AccountID, sqlutil.FormatTime(i.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert gift idea: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete implements usecase.Repository; an idea of someone else is left
// alone without an error, so its existence stays private.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	if _, err := r.db.ExecContext(ctx, `DELETE FROM gifts_ideas WHERE id = ? AND `+owned, append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete gift idea: %w", err)
	}
	return nil
}
