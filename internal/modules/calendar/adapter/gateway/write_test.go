package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
)

func TestFormatEventRoundTrip(t *testing.T) {
	start := time.Date(2026, 11, 12, 9, 30, 0, 0, time.UTC)
	e, _ := domain.NewEvent("abc@trisse-home", "Zahnarzt; Kontrolle, Teil 2 — "+strings.Repeat("ä", 40), "Main St\n2nd floor", start, start.Add(time.Hour), false)
	ics := FormatEvent(e, start)
	for _, l := range strings.Split(ics, "\r\n") {
		if len(l) > 75 {
			t.Fatalf("line longer than 75 octets: %q", l)
		}
	}
	parsed, err := ParseICalendar(ics, time.UTC)
	if err != nil || len(parsed) != 1 {
		t.Fatalf("parse: %v %v", parsed, err)
	}
	if parsed[0].Summary != e.Summary || parsed[0].Location != e.Location || !parsed[0].Start.Equal(start) {
		t.Fatalf("round trip lost data: %+v", parsed[0])
	}
	day := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	allDay, _ := domain.NewEvent("x", "Holiday", "", day, day.AddDate(0, 0, 1), true)
	if ics := FormatEvent(allDay, start); !strings.Contains(ics, "DTSTART;VALUE=DATE:20261120\r\n") || !strings.Contains(ics, "DTEND;VALUE=DATE:20261121") {
		t.Fatalf("all-day format: %s", ics)
	}
}

func TestCreate(t *testing.T) {
	var gotPath, gotBody, ifNoneMatch string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, _ := r.BasicAuth(); u != "me" || p != "pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(r.Body)
		gotPath, gotBody, ifNoneMatch = r.URL.Path, string(b), r.Header.Get("If-None-Match")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	c := NewCalDAV(srv.Client(), time.UTC)
	s := domain.Settings{Username: "me", Password: "pw"}
	start := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
	e, _ := domain.NewEvent("u1@trisse-home", "Test", "", start, start.Add(time.Hour), false)
	if err := c.Create(context.Background(), s, srv.URL+"/cal/home", e); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/cal/home/u1-trisse-home.ics" || ifNoneMatch != "*" || !strings.Contains(gotBody, "SUMMARY:Test") {
		t.Fatalf("request %q %q %q", gotPath, ifNoneMatch, gotBody)
	}
	s.Password = "wrong"
	if err := c.Create(context.Background(), s, srv.URL+"/cal/home/", e); err == nil {
		t.Fatal("expected error")
	}
}
