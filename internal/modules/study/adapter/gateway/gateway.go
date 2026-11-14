// Package gateway stores the study modules in SQLite.
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

	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are included in data exports.
var TransferTables = []sqlutil.Table{{Name: "study_settings"}, {Name: "study_modules"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func scan(row interface{ Scan(...any) error }) (domain.Module, error) {
	var m domain.Module
	var status string
	var grade sql.NullFloat64
	if err := row.Scan(&m.ID, &m.Name, &m.ECTS, &m.Semester, &status, &grade); err != nil {
		return domain.Module{}, err
	}
	m.Status = domain.Status(status)
	if grade.Valid {
		g := grade.Float64
		m.Grade = &g
	}
	return m, nil
}

const selectModules = `SELECT id, name, ects, semester, status, grade FROM study_modules`

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Module, error) {
	rows, err := r.db.QueryContext(ctx, selectModules)
	if err != nil {
		return nil, fmt.Errorf("select modules: %w", err)
	}
	defer rows.Close()
	out := []domain.Module{}
	for rows.Next() {
		m, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Module, error) {
	m, err := scan(r.db.QueryRowContext(ctx, selectModules+` WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Module{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Module{}, fmt.Errorf("select module: %w", err)
	}
	return m, nil
}

// Save implements usecase.Repository (insert or update).
func (r *Repository) Save(ctx context.Context, m domain.Module) error {
	var grade sql.NullFloat64
	if m.Grade != nil {
		grade = sql.NullFloat64{Float64: *m.Grade, Valid: true}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO study_modules (id, name, ects, semester, status, grade) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, ects = excluded.ects, semester = excluded.semester,
			status = excluded.status, grade = excluded.grade`, m.ID, m.Name, m.ECTS, m.Semester, string(m.Status), grade)
	if err != nil {
		return fmt.Errorf("upsert module: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM study_modules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete module: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// TotalECTS implements usecase.Repository.
func (r *Repository) TotalECTS(ctx context.Context) (float64, error) {
	var v float64
	err := r.db.QueryRowContext(ctx, `SELECT total_ects FROM study_settings WHERE id = 1`).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultTotalECTS, nil
	}
	if err != nil {
		return 0, fmt.Errorf("select total: %w", err)
	}
	return v, nil
}

// SetTotalECTS implements usecase.Repository.
func (r *Repository) SetTotalECTS(ctx context.Context, v float64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO study_settings (id, total_ects) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET total_ects = excluded.total_ects`, v)
	if err != nil {
		return fmt.Errorf("upsert total: %w", err)
	}
	return nil
}
