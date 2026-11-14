package gateway

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestICSSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nUID:1\r\nSUMMARY:Biotonne\r\nDTSTART;VALUE=DATE:20261112\r\n"+
			"RRULE:FREQ=WEEKLY;INTERVAL=2\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nUID:2\r\nSUMMARY:Papier\r\n"+
			"DTSTART;TZID=Europe/Berlin:20261113T063000\r\nDTEND;TZID=Europe/Berlin:20261113T070000\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n")
	}))
	defer srv.Close()
	loc, _ := time.LoadLocation("Europe/Berlin")
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	src := domain.Source{Kind: domain.SourceURL, URL: srv.URL}
	cs, err := NewICSSource(srv.Client(), nil, loc).Collections(context.Background(), src, from, from.AddDate(0, 1, 0))
	if err != nil || len(cs) != 3 {
		t.Fatalf("%+v %v", cs, err)
	}
	if cs[2].Summary != "Papier" || cs[2].Date.Day() != 13 {
		t.Fatalf("timed event date: %+v", cs[2])
	}
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepository(dbtest.Open(t, "waste", Migrations()))
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatal(err)
	}
	_ = repo.Save(ctx, domain.Settings{Source: domain.Source{Kind: domain.SourceURL, URL: "https://x/a.ics"}, Rules: []domain.Rule{{Match: "bio", Icon: "🍂"}}})
	if s, err := repo.Load(ctx); err != nil || s.Rules[0].Icon != "🍂" || s.Source.Kind != domain.SourceURL || s.Source.URL != "https://x/a.ics" {
		t.Fatalf("%+v %v", s, err)
	}
}

// TestURLRowsFromBeforeTheMigration keeps settings saved before #153 working.
func TestURLRowsFromBeforeTheMigration(t *testing.T) {
	db := dbtest.Open(t, "waste", Migrations())
	if _, err := db.Exec(`INSERT INTO waste_settings (id, url, rules) VALUES (1, 'https://x/old.ics', '[]')`); err != nil {
		t.Fatal(err)
	}
	s, err := NewSettingsRepository(db).Load(context.Background())
	if err != nil || s.Source.Kind != domain.SourceURL || s.Source.URL != "https://x/old.ics" {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestFileSource(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "waste", Migrations())
	repo := NewSettingsRepository(db)
	loc, _ := time.LoadLocation("Europe/Berlin")
	content, err := os.ReadFile("testdata/abfuhrkalender-2027.ics")
	if err != nil {
		t.Fatal(err)
	}

	// The parser reads a real-world municipal file: VTIMEZONE, all-day
	// series with UNTIL and EXDATE, VALARM, folded descriptions, a zoned
	// time and a UTC time that is already the next day in Berlin.
	from, to := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC)
	cs, err := NewParser(loc).Parse(content, from, to)
	if err != nil {
		t.Fatal(err)
	}
	count := map[string]int{}
	for _, c := range cs {
		count[c.Summary]++
	}
	// Every second Tuesday from 5 Jan 2027 to the end of the year is 26 dates, minus the EXDATE.
	if count["Restmülltonne"] != 25 || count["Biotonne"] != 1 || count["Papier / Pappe"] != 1 || count["Gelber Sack"] != 1 {
		t.Fatalf("counts: %v", count)
	}
	for _, c := range cs {
		if c.Summary == "Gelber Sack" && c.Date.Format(time.DateOnly) != "2027-02-01" {
			t.Fatalf("UTC 23:30 is 1 February in Berlin: %v", c.Date)
		}
	}

	day := time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)
	src, _ := domain.NewFileSource("abfuhrkalender-2027.ics", time.Date(2026, 12, 20, 9, 0, 0, 0, time.UTC), cs)
	s := domain.Settings{Source: src, Rules: []domain.Rule{{Match: "bio", Label: "Organic"}}}
	if err := repo.SaveWithFile(ctx, s, content); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Load(ctx)
	if err != nil || got.Source.Kind != domain.SourceFile || got.Source.File != src.File {
		t.Fatalf("round trip: %+v vs %+v (%v)", got.Source.File, src.File, err)
	}

	// The file source reads the stored content; there is no HTTP client at all.
	stored, err := NewICSSource(nil, db, loc).Collections(ctx, got.Source, day, day.AddDate(0, 0, 1))
	if err != nil || len(stored) != 1 || stored[0].Summary != "Restmülltonne" {
		t.Fatalf("stored: %+v %v", stored, err)
	}

	// New rules without content keep the file; a URL source deletes it.
	got.Rules = nil
	if err := repo.SaveWithFile(ctx, got, nil); err != nil || fileCount(t, db) != 1 {
		t.Fatalf("rules only: %v", err)
	}
	if err := repo.Save(ctx, domain.Settings{Source: domain.Source{Kind: domain.SourceURL, URL: "https://x/a.ics"}}); err != nil || fileCount(t, db) != 0 {
		t.Fatalf("switch to URL: %v", err)
	}
	if got, _ := repo.Load(ctx); got.Source.File != (domain.CalendarFile{}) {
		t.Fatalf("file metadata must be cleared: %+v", got.Source.File)
	}
	if _, err := NewICSSource(nil, db, loc).Collections(ctx, src, day, day); err == nil {
		t.Fatal("a missing file must be an error")
	}
}

func fileCount(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM waste_calendar_file`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
