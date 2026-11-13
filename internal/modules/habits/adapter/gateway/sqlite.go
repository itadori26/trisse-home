// Package gateway stores habits and their completions in SQLite.
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
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/habits/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports (habits first,
// so completions can reference them on import).
var TransferTables = []sqlutil.Table{{Name: "habits"}, {Name: "habit_completions"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, name, emoji, days, position, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Habit, error) {
	var h domain.Habit
	var days, created string
	if err := sc.Scan(&h.ID, &h.Name, &h.Emoji, &days, &h.Position, &created); err != nil {
		return h, err
	}
	var nums []int
	for _, p := range strings.Split(days, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			nums = append(nums, n)
		}
	}
	var err error
	if h.Schedule, err = domain.NewSchedule(nums); err != nil {
		return h, err
	}
	h.CreatedAt, err = sqlutil.ParseTime(created)
	return h, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Habit, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM habits ORDER BY position, created_at`)
	if err != nil {
		return nil, fmt.Errorf("query habits: %w", err)
	}
	defer rows.Close()
	var out []domain.Habit
	for rows.Next() {
		h, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan habit: %w", err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Habit, error) {
	h, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM habits WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Habit{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Habit{}, fmt.Errorf("select habit: %w", err)
	}
	return h, nil
}

// Save implements usecase.Repository (upsert).
func (r *Repository) Save(ctx context.Context, h domain.Habit) error {
	days := make([]string, 0, 7)
	for _, d := range h.Schedule.Days() {
		days = append(days, strconv.Itoa(d))
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO habits (`+columns+`) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, emoji = excluded.emoji, days = excluded.days, position = excluded.position`,
		h.ID, h.Name, h.Emoji, strings.Join(days, ","), h.Position, sqlutil.FormatTime(h.CreatedAt))
	if err != nil {
		return fmt.Errorf("upsert habit: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository; completions go with the habit.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM habits WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete habit: %w", err)
	}
	return nil
}

// Completions implements usecase.Repository.
func (r *Repository) Completions(ctx context.Context, since domain.Day) (map[string]map[domain.Day]bool, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT habit_id, day FROM habit_completions WHERE day >= ?`, since.Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("query completions: %w", err)
	}
	defer rows.Close()
	out := map[string]map[domain.Day]bool{}
	for rows.Next() {
		var id, day string
		if err := rows.Scan(&id, &day); err != nil {
			return nil, fmt.Errorf("scan completion: %w", err)
		}
		d, err := time.Parse(time.DateOnly, day)
		if err != nil {
			continue // ignore a malformed imported row
		}
		if out[id] == nil {
			out[id] = map[domain.Day]bool{}
		}
		out[id][d] = true
	}
	return out, rows.Err()
}

// SetCompletion implements usecase.Repository.
func (r *Repository) SetCompletion(ctx context.Context, id string, day domain.Day, done bool) error {
	q := `DELETE FROM habit_completions WHERE habit_id = ? AND day = ?`
	if done {
		q = `INSERT OR IGNORE INTO habit_completions (habit_id, day) VALUES (?, ?)`
	}
	if _, err := r.db.ExecContext(ctx, q, id, day.Format(time.DateOnly)); err != nil {
		return fmt.Errorf("set completion: %w", err)
	}
	return nil
}
