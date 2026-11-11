package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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
	cs, err := NewICSSource(srv.Client(), loc).Collections(context.Background(), srv.URL, from, from.AddDate(0, 1, 0))
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
	_ = repo.Save(ctx, domain.Settings{URL: "https://x/a.ics", Rules: []domain.Rule{{Match: "bio", Icon: "🍂"}}})
	if s, err := repo.Load(ctx); err != nil || s.Rules[0].Icon != "🍂" {
		t.Fatalf("%+v %v", s, err)
	}
}
