package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/onboarding/domain"
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

// Load returns the stored state or the pending zero state.
func (r *Repository) Load(ctx context.Context) (domain.State, error) {
	var completed *string
	err := r.db.QueryRowContext(ctx, `SELECT completed_at FROM onboarding_state WHERE id = 1`).Scan(&completed)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.State{}, nil
	}
	if err != nil {
		return domain.State{}, fmt.Errorf("select onboarding state: %w", err)
	}
	at, err := sqlutil.ParseNullableTime(completed)
	if err != nil {
		return domain.State{}, err
	}
	return domain.RestoreState(at), nil
}

// Save stores the state.
func (r *Repository) Save(ctx context.Context, s domain.State) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO onboarding_state (id, completed_at) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET completed_at = excluded.completed_at`, sqlutil.FormatNullableTime(s.CompletedAt()))
	if err != nil {
		return fmt.Errorf("upsert onboarding state: %w", err)
	}
	return nil
}
