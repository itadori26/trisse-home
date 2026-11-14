// Package gateway stores the settings and the school holiday cache in SQLite
// and loads school holidays from an ICS link.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/shared/ical"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are exported (the cache is loaded again after an import).
var TransferTables = []sqlutil.Table{{Name: "holidays_settings"}}

// Store implements usecase.SettingsRepository and usecase.SchoolCache.
type Store struct{ db *sql.DB }

// NewStore returns a store backed by db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Load implements usecase.SettingsRepository; no row means no link.
func (s *Store) Load(ctx context.Context) (domain.Settings, error) {
	var out domain.Settings
	err := s.db.QueryRowContext(ctx, `SELECT school_url FROM holidays_settings WHERE id = 1`).Scan(&out.SchoolURL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	return out, nil
}

// Save implements usecase.SettingsRepository.
func (s *Store) Save(ctx context.Context, st domain.Settings) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO holidays_settings (id, school_url) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET school_url = excluded.school_url`, st.SchoolURL)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

// LoadSchool implements usecase.SchoolCache.
func (s *Store) LoadSchool(ctx context.Context) ([]domain.SchoolHoliday, string, time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, start_date, end_date, url, loaded_at FROM holidays_school ORDER BY start_date, name`)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("select school holidays: %w", err)
	}
	defer rows.Close()
	var (
		out []domain.SchoolHoliday
		url string
		at  time.Time
	)
	for rows.Next() {
		var h domain.SchoolHoliday
		var start, end, loaded string
		if err := rows.Scan(&h.Name, &start, &end, &url, &loaded); err != nil {
			return nil, "", time.Time{}, fmt.Errorf("scan school holiday: %w", err)
		}
		h.Start, _ = time.Parse(time.DateOnly, start) // written by StoreSchool
		h.End, _ = time.Parse(time.DateOnly, end)
		at, _ = time.Parse(time.RFC3339, loaded)
		out = append(out, h)
	}
	return out, url, at, rows.Err()
}

// StoreSchool implements usecase.SchoolCache: replaces the cache, one row
// per period with the year it starts in.
func (s *Store) StoreSchool(ctx context.Context, url string, periods []domain.SchoolHoliday, at time.Time) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `DELETE FROM holidays_school`); err != nil {
		return fmt.Errorf("clear school holidays: %w", err)
	}
	for _, p := range periods {
		if _, err = tx.ExecContext(ctx, `INSERT INTO holidays_school (year, name, start_date, end_date, url, loaded_at) VALUES (?, ?, ?, ?, ?, ?)`,
			p.Start.Year(), p.Name, p.Start.Format(time.DateOnly), p.End.Format(time.DateOnly), url, at.UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("insert school holiday: %w", err)
		}
	}
	return tx.Commit()
}

// ClearSchool implements usecase.SchoolCache.
func (s *Store) ClearSchool(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM holidays_school`)
	return err
}

// Cache adapts Store to usecase.SchoolCache (whose methods are named Load,
// Store and Clear).
type Cache struct{ s *Store }

// Load implements usecase.SchoolCache.
func (c Cache) Load(ctx context.Context) ([]domain.SchoolHoliday, string, time.Time, error) {
	return c.s.LoadSchool(ctx)
}

// Store implements usecase.SchoolCache.
func (c Cache) Store(ctx context.Context, url string, p []domain.SchoolHoliday, at time.Time) error {
	return c.s.StoreSchool(ctx, url, p, at)
}

// Clear implements usecase.SchoolCache.
func (c Cache) Clear(ctx context.Context) error { return c.s.ClearSchool(ctx) }

// NewCache returns the school holiday cache of store.
func NewCache(store *Store) Cache { return Cache{s: store} }

const maxICSBytes = 2 << 20

// ICSSource implements usecase.SchoolSource with the shared ICS parser.
type ICSSource struct {
	client *http.Client
	loc    *time.Location
}

// NewICSSource returns a source; floating times are interpreted in loc.
func NewICSSource(client *http.Client, loc *time.Location) *ICSSource {
	return &ICSSource{client: client, loc: loc}
}

// Fetch implements usecase.SchoolSource: every event (all-day or timed) is a
// period from its first to its last day; recurrences are expanded.
func (s *ICSSource) Fetch(ctx context.Context, url string, from, to time.Time) ([]domain.SchoolHoliday, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/calendar, */*;q=0.5")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server answered %s", res.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxICSBytes))
	if err != nil {
		return nil, fmt.Errorf("read calendar: %w", err)
	}
	return Parse(string(raw), s.loc, from, to)
}

// Parse turns ICS events into school holiday periods (End inclusive).
func Parse(data string, loc *time.Location, from, to time.Time) ([]domain.SchoolHoliday, error) {
	events, err := ical.Parse(data, loc)
	if err != nil {
		return nil, err
	}
	var out []domain.SchoolHoliday
	for _, e := range events {
		for _, occ := range ical.Expand(e, from, to) {
			start, end := occ.Start, occ.End
			if !occ.AllDay {
				start, end = start.In(loc), end.In(loc)
			}
			last := domain.Date(end)
			if occ.AllDay || (end.Hour() == 0 && end.Minute() == 0 && end.Second() == 0) {
				last = last.AddDate(0, 0, -1) // the end is exclusive
			}
			first := domain.Date(start)
			if last.Before(first) {
				last = first
			}
			out = append(out, domain.SchoolHoliday{Name: strings.TrimSpace(occ.Summary), Start: first, End: last})
		}
	}
	slices.SortFunc(out, func(a, b domain.SchoolHoliday) int { return a.Start.Compare(b.Start) })
	return out, nil
}
