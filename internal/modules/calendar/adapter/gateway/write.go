package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/ical"
)

const (
	icalDateLayout     = "20060102"
	icalDateTimeLayout = "20060102T150405Z"
)

// FormatEvent serialises e as a VCALENDAR with one VEVENT.
func FormatEvent(e domain.Event, stamp time.Time) string {
	lines := []string{
		"BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//trisse-home//calendar//EN", "CALSCALE:GREGORIAN",
		"BEGIN:VEVENT",
		"UID:" + e.UID,
		"DTSTAMP:" + stamp.UTC().Format(icalDateTimeLayout),
	}
	if e.AllDay {
		lines = append(lines, "DTSTART;VALUE=DATE:"+e.Start.Format(icalDateLayout), "DTEND;VALUE=DATE:"+e.End.Format(icalDateLayout))
	} else {
		lines = append(lines, "DTSTART:"+e.Start.UTC().Format(icalDateTimeLayout), "DTEND:"+e.End.UTC().Format(icalDateTimeLayout))
	}
	lines = append(lines, "SUMMARY:"+ical.EscapeText(e.Summary))
	if e.Location != "" {
		lines = append(lines, "LOCATION:"+ical.EscapeText(e.Location))
	}
	lines = append(lines, "END:VEVENT", "END:VCALENDAR")
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(ical.FoldLine(l))
	}
	return b.String()
}

// Create implements usecase.EventWriter with a CalDAV PUT. If-None-Match
// makes sure an existing object is never overwritten.
func (c *CalDAV) Create(ctx context.Context, s domain.Settings, calendarURL string, e domain.Event) error {
	if !strings.HasSuffix(calendarURL, "/") {
		calendarURL += "/"
	}
	target := calendarURL + url.PathEscape(strings.ReplaceAll(e.UID, "@", "-")) + ".ics"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, strings.NewReader(FormatEvent(e, time.Now())))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(s.Username, s.Password)
	req.Header.Set("Content-Type", "text/calendar; charset=utf-8")
	req.Header.Set("If-None-Match", "*")
	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server answered %s", res.Status)
	}
	return nil
}
