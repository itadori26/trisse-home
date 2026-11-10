package gateway

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"trisse-home/internal/modules/calendar/domain"
)

const (
	maxCalDAVResponseBytes = 8 << 20
	caldavTimeLayout       = "20060102T150405Z"
)

// CalDAV loads events with a calendar-query REPORT. The server expands
// recurring events into single instances (RFC 4791 §9.6.5), so the parser
// does not need to implement RRULE.
type CalDAV struct {
	client *http.Client
	loc    *time.Location
}

// NewCalDAV returns a client; floating times are interpreted in loc.
func NewCalDAV(client *http.Client, loc *time.Location) *CalDAV {
	return &CalDAV{client: client, loc: loc}
}

const reportTemplate = `<?xml version="1.0" encoding="utf-8"?>
<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>
    <C:calendar-data><C:expand start="%[1]s" end="%[2]s"/></C:calendar-data>
  </D:prop>
  <C:filter>
    <C:comp-filter name="VCALENDAR">
      <C:comp-filter name="VEVENT"><C:time-range start="%[1]s" end="%[2]s"/></C:comp-filter>
    </C:comp-filter>
  </C:filter>
</C:calendar-query>`

type multistatus struct {
	Responses []struct {
		Propstats []struct {
			CalendarData string `xml:"prop>calendar-data"`
			Status       string `xml:"status"`
		} `xml:"propstat"`
	} `xml:"response"`
}

// Events implements usecase.EventSource.
func (c *CalDAV) Events(ctx context.Context, s domain.Settings, calendarURL string, from, to time.Time) ([]domain.Event, error) {
	body := fmt.Sprintf(reportTemplate, from.UTC().Format(caldavTimeLayout), to.UTC().Format(caldavTimeLayout))
	req, err := http.NewRequestWithContext(ctx, "REPORT", calendarURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(s.Username, s.Password)
	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("server answered %s", res.Status)
	}
	var ms multistatus
	if err := xml.NewDecoder(io.LimitReader(res.Body, maxCalDAVResponseBytes)).Decode(&ms); err != nil {
		return nil, fmt.Errorf("parse CalDAV response: %w", err)
	}
	name := domain.CalendarName(calendarURL)
	var events []domain.Event
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if ps.CalendarData == "" {
				continue
			}
			parsed, err := ParseICalendar(ps.CalendarData, c.loc)
			if err != nil {
				continue // one broken object should not hide the rest of the calendar
			}
			for _, e := range parsed {
				e.Calendar = name
				events = append(events, e)
			}
		}
	}
	return events, nil
}
