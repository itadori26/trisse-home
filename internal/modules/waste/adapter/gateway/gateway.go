// Package gateway stores the waste settings and loads ICS subscriptions.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/shared/ical"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

const maxICSBytes = 4 << 20

// SettingsRepository stores the settings.
type SettingsRepository struct{ db *sql.DB }

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB) *SettingsRepository { return &SettingsRepository{db: db} }

type ruleRow struct {
	Match string `json:"match"`
	Label string `json:"label"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// Load implements usecase.SettingsRepository.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var raw, kind, uploadedAt, from, to string
	err := r.db.QueryRowContext(ctx, `SELECT url, rules, source, file_name, file_uploaded_at, file_events, file_from, file_to
		FROM waste_settings WHERE id = 1`).Scan(&s.Source.URL, &raw, &kind, &s.Source.File.Name, &uploadedAt, &s.Source.File.Events, &from, &to)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	s.Source.Kind = domain.SourceKind(kind)
	if s.Source.Kind == domain.SourceFile {
		s.Source.File.UploadedAt, _ = time.Parse(time.RFC3339, uploadedAt) // written by Save
		s.Source.File.From, _ = time.Parse(time.DateOnly, from)
		s.Source.File.To, _ = time.Parse(time.DateOnly, to)
	} else {
		s.Source.File = domain.CalendarFile{}
	}
	var rows []ruleRow
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return domain.Settings{}, fmt.Errorf("decode rules: %w", err)
	}
	s.Rules = make([]domain.Rule, 0, len(rows))
	for _, x := range rows {
		s.Rules = append(s.Rules, domain.Rule(x))
	}
	return s, nil
}

// Save implements usecase.SettingsRepository; a stored file is deleted.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	return r.save(ctx, s, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `DELETE FROM waste_calendar_file`)
		return err
	})
}

// SaveWithFile implements usecase.SettingsRepository.
func (r *SettingsRepository) SaveWithFile(ctx context.Context, s domain.Settings, content []byte) error {
	return r.save(ctx, s, func(tx *sql.Tx) error {
		if content == nil {
			return nil
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO waste_calendar_file (id, content) VALUES (1, ?)
			ON CONFLICT (id) DO UPDATE SET content = excluded.content`, string(content))
		return err
	})
}

func (r *SettingsRepository) save(ctx context.Context, s domain.Settings, file func(*sql.Tx) error) (err error) {
	rows := make([]ruleRow, 0, len(s.Rules))
	for _, x := range s.Rules {
		rows = append(rows, ruleRow(x))
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode rules: %w", err)
	}
	f := s.Source.File
	var uploadedAt, from, to string
	if s.Source.Kind == domain.SourceFile {
		uploadedAt, from, to = f.UploadedAt.UTC().Format(time.RFC3339), f.From.Format(time.DateOnly), f.To.Format(time.DateOnly)
	} else {
		f = domain.CalendarFile{}
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `INSERT INTO waste_settings (id, url, rules, source, file_name, file_uploaded_at, file_events, file_from, file_to)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, rules = excluded.rules, source = excluded.source,
			file_name = excluded.file_name, file_uploaded_at = excluded.file_uploaded_at, file_events = excluded.file_events,
			file_from = excluded.file_from, file_to = excluded.file_to`,
		s.Source.URL, string(raw), string(s.Source.Kind), f.Name, uploadedAt, f.Events, from, to); err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	if err = file(tx); err != nil {
		return fmt.Errorf("store calendar file: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ICSSource implements usecase.CollectionSource: subscriptions are fetched,
// uploaded files are read from the database without any outbound request.
type ICSSource struct {
	client *http.Client
	db     *sql.DB
	parser Parser
}

// NewICSSource returns a source; floating times are interpreted in loc.
func NewICSSource(client *http.Client, db *sql.DB, loc *time.Location) *ICSSource {
	return &ICSSource{client: client, db: db, parser: Parser{loc: loc}}
}

// Collections implements usecase.CollectionSource.
func (s *ICSSource) Collections(ctx context.Context, src domain.Source, from, to time.Time) ([]domain.Collection, error) {
	switch src.Kind {
	case domain.SourceURL:
		raw, err := s.fetch(ctx, src.URL)
		if err != nil {
			return nil, err
		}
		return s.parser.Parse(raw, from, to)
	case domain.SourceFile:
		var content string
		err := s.db.QueryRowContext(ctx, `SELECT content FROM waste_calendar_file WHERE id = 1`).Scan(&content)
		if err != nil {
			return nil, fmt.Errorf("read stored calendar: %w", err)
		}
		return s.parser.Parse([]byte(content), from, to)
	default:
		return nil, domain.ErrNotConfigured
	}
}

func (s *ICSSource) fetch(ctx context.Context, url string) ([]byte, error) {
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
	return raw, nil
}

// Parser implements usecase.CalendarParser with the shared ICS parser.
type Parser struct{ loc *time.Location }

// NewParser returns a parser; floating times are interpreted in loc.
func NewParser(loc *time.Location) Parser { return Parser{loc: loc} }

// Parse implements usecase.CalendarParser: the dates of all events between
// from and to, recurrences expanded.
func (p Parser) Parse(content []byte, from, to time.Time) ([]domain.Collection, error) {
	events, err := ical.Parse(string(content), p.loc)
	if err != nil {
		return nil, err
	}
	var out []domain.Collection
	for _, e := range events {
		for _, occ := range ical.Expand(e, from, to) {
			date := occ.Start
			if !occ.AllDay {
				date = date.In(p.loc)
			}
			out = append(out, domain.Collection{Date: domain.Date(date), Summary: occ.Summary})
		}
	}
	return out, nil
}
