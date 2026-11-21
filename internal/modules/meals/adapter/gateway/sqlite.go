// Package gateway stores the meal plan in SQLite.
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

	"waylight/internal/modules/meals/domain"
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
var TransferTables = []sqlutil.Table{{Name: "meals_entries"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, day, meal, dish, link, ingredients, owner_id, visibility, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Entry, error) {
	var e domain.Entry
	var day, meal, ingredients, owner, vis, created string
	if err := sc.Scan(&e.ID, &day, &meal, &e.Dish, &e.Link, &ingredients, &owner, &vis, &created); err != nil {
		return e, err
	}
	var err error
	if e.Day, err = time.Parse(time.DateOnly, day); err != nil {
		return e, err
	}
	e.Meal = domain.Meal(meal)
	if ingredients != "" {
		e.Ingredients = strings.Split(ingredients, "\n")
	}
	e.Access = kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}
	e.CreatedAt, err = sqlutil.ParseTime(created)
	return e, err
}

// Between implements usecase.Repository: the visible dishes of the days
// from (inclusive) to (exclusive).
func (r *Repository) Between(ctx context.Context, from, to time.Time) ([]domain.Entry, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM meals_entries WHERE day >= ? AND day < ? AND `+cond+` ORDER BY day, created_at`,
		append([]any{from.Format(time.DateOnly), to.Format(time.DateOnly)}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("query dishes: %w", err)
	}
	defer rows.Close()
	var out []domain.Entry
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dish: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Entry, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	e, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM meals_entries WHERE id = ? AND `+cond, append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Entry{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Entry{}, fmt.Errorf("select dish: %w", err)
	}
	return e, nil
}

// Save implements usecase.Repository (upsert). A new dish belongs to the
// principal whatever e.Access says; an existing one is changed only by its
// owner (ADR 0018).
func (r *Repository) Save(ctx context.Context, e domain.Entry) error {
	p := kernel.PrincipalFrom(ctx)
	owned, ownedArgs := sqlutil.OwnedBy(p, "meals_entries")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO meals_entries (id, day, meal, dish, link, ingredients, owner_id, visibility, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET day = excluded.day, meal = excluded.meal, dish = excluded.dish, link = excluded.link,
		ingredients = excluded.ingredients, visibility = excluded.visibility
		WHERE `+owned,
		append([]any{e.ID, e.Date(), string(e.Meal), e.Dish, e.Link, strings.Join(e.Ingredients, "\n"),
			p.AccountID, string(e.Access.Visibility), sqlutil.FormatTime(e.CreatedAt)}, ownedArgs...)...)
	if err != nil {
		return fmt.Errorf("upsert dish: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return kernel.ErrNotOwner
	}
	return nil
}

// Delete implements usecase.Repository for the owner; a visible dish of
// someone else returns kernel.ErrNotOwner, an unknown one is no error.
func (r *Repository) Delete(ctx context.Context, id string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // the condition comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `DELETE FROM meals_entries WHERE id = ? AND `+owned, append([]any{id}, args...)...)
	if err != nil {
		return fmt.Errorf("delete dish: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := r.Get(ctx, id); err == nil {
			return kernel.ErrNotOwner
		}
	}
	return nil
}
