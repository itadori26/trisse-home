package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/platform/database/dbtest"
)

const sample = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n" +
	"BEGIN:VEVENT\r\nUID:a\r\nSUMMARY:Team meeting\\, weekly\r\nLOCATION:Room 1\r\nDTSTART;TZID=Europe/Berlin:20261110T090000\r\nDTEND;TZID=Europe/Berlin:20261110T100000\r\n" +
	"BEGIN:VALARM\r\nTRIGGER:-PT15M\r\nDESCRIPTION:Alarm summary must be ignored\r\nEND:VALARM\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:b\r\nSUMMARY:Holiday\r\nDTSTART;VALUE=DATE:20261224\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:c\r\nSUMMARY:Call with a very long folded\r\n  title\r\nDTSTART:20261111T150000Z\r\nDURATION:PT1H30M\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:d\r\nSUMMARY:Cancelled\r\nSTATUS:CANCELLED\r\nDTSTART:20261112T150000Z\r\nEND:VEVENT\r\n" +
	"END:VCALENDAR\r\n"

func TestParseICalendar(t *testing.T) {
	events, err := ParseICalendar(sample, time.UTC)
	if err != nil || len(events) != 3 {
		t.Fatalf("got %d events, err %v", len(events), err)
	}
	meeting := events[0]
	if meeting.Summary != "Team meeting, weekly" || meeting.Location != "Room 1" || meeting.Start.UTC().Hour() != 8 || meeting.End.Sub(meeting.Start) != time.Hour {
		t.Fatalf("meeting parsed wrong: %+v", meeting)
	}
	holiday := events[1]
	if !holiday.AllDay || holiday.End.Sub(holiday.Start) != 24*time.Hour {
		t.Fatalf("all-day event wrong: %+v", holiday)
	}
	call := events[2]
	if call.Summary != "Call with a very long folded title" || call.End.Sub(call.Start) != 90*time.Minute {
		t.Fatalf("folding/duration wrong: %+v", call)
	}
	if _, err := ParseICalendar("garbage", time.UTC); err == nil {
		t.Fatal("expected error for non-iCalendar input")
	}
}

func TestCalDAV(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "REPORT" || r.Header.Get("Depth") != "1" || !strings.Contains(string(body), "<C:expand") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, p, _ := r.BasicAuth(); p != "pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><d:multistatus xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav">
			<d:response><d:href>/x.ics</d:href><d:propstat><d:prop><cal:calendar-data>` + sample + `</cal:calendar-data></d:prop>
			<d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response></d:multistatus>`))
	}))
	defer srv.Close()
	c := NewCalDAV(srv.Client(), time.UTC)
	s := domain.Settings{Username: "me", Password: "pw"}
	events, err := c.Events(context.Background(), s, srv.URL+"/calendars/me/personal/", time.Now(), time.Now().Add(time.Hour))
	if err != nil || len(events) != 3 || events[0].Calendar != "personal" {
		t.Fatalf("unexpected %v %v", events, err)
	}
	s.Password = "wrong"
	if _, err := c.Events(context.Background(), s, srv.URL+"/x/", time.Now(), time.Now()); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401, got %v", err)
	}
}

type plainCipher struct{}

func (plainCipher) Encrypt(s string) (string, error) { return "x" + s, nil }
func (plainCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "x") {
		return "", errors.New("bad")
	}
	return s[1:], nil
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepository(dbtest.Open(t, "calendar", Migrations()), plainCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s := domain.Settings{CalendarURLs: []string{"https://a/", "https://b/"}, Username: "me", Password: "pw"}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Load(ctx)
	if err != nil || len(got.CalendarURLs) != 2 || got.Password != "pw" {
		t.Fatalf("round trip %+v %v", got, err)
	}
}
