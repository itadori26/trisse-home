package gateway

import (
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/ical"
)

// ParseICalendar extracts the VEVENTs of an iCalendar object as domain
// events. CalDAV servers expand recurrences (calendar-query with expand), so
// no expansion happens here.
func ParseICalendar(data string, loc *time.Location) ([]domain.Event, error) {
	parsed, err := ical.Parse(data, loc)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Event, 0, len(parsed))
	for _, e := range parsed {
		out = append(out, domain.Event{UID: e.UID, Summary: e.Summary, Location: e.Location, Description: e.Description,
			URL: domain.SafeLink(e.URL), Start: e.Start, End: e.End, AllDay: e.AllDay})
	}
	return out, nil
}
