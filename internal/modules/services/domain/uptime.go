package domain

import "time"

// Uptime retention and debouncing (#81).
const (
	AggregateRetention = 90 * 24 * time.Hour
	// IncidentConfirmations matches the notification debouncing.
	IncidentConfirmations = 2
)

// DailyAggregate counts the checks of one service on one day; checks during
// maintenance are counted separately and excluded from the uptime.
type DailyAggregate struct {
	ServiceID   string
	Day         string // "2006-01-02" in the server's time zone
	Checks      int
	Online      int
	Maintenance int
}

// DayOf formats the aggregate key of t.
func DayOf(t time.Time) string { return t.Format(time.DateOnly) }

// Uptime returns the share of online checks in percent and whether there was
// anything to measure.
func Uptime(aggs []DailyAggregate) (float64, bool) {
	var counted, online int
	for _, a := range aggs {
		counted += a.Checks - a.Maintenance
		online += a.Online
	}
	if counted <= 0 {
		return 0, false
	}
	return min(100, 100*float64(online)/float64(counted)), true
}

// Incident is a confirmed outage of a service.
type Incident struct {
	ID        string
	ServiceID string
	Start     time.Time
	// End is zero while the incident is open.
	End   time.Time
	Cause string
}

// Open reports whether the incident is still ongoing.
func (i Incident) Open() bool { return i.End.IsZero() }

// Duration returns the length, up to now for open incidents.
func (i Incident) Duration(now time.Time) time.Duration {
	if i.Open() {
		return now.Sub(i.Start)
	}
	return i.End.Sub(i.Start)
}

// IncidentTrack debounces check results: a state change is confirmed after
// IncidentConfirmations consecutive checks and dated to the first of them.
type IncidentTrack struct {
	Known     bool
	Online    bool
	candidate bool
	count     int
	since     time.Time
	cause     string
}

// Transition is a confirmed change of an IncidentTrack.
type Transition struct {
	Changed bool
	Online  bool
	At      time.Time
	Cause   string
}

// Advance feeds one check result into the track.
func (t IncidentTrack) Advance(online bool, at time.Time, cause string) (IncidentTrack, Transition) {
	if !t.Known {
		return IncidentTrack{Known: true, Online: online}, Transition{}
	}
	if online == t.Online {
		t.count = 0
		return t, Transition{}
	}
	if t.count == 0 || t.candidate != online {
		t.candidate, t.count, t.since, t.cause = online, 0, at, cause
	}
	t.count++
	if t.count < IncidentConfirmations {
		return t, Transition{}
	}
	return IncidentTrack{Known: true, Online: online}, Transition{Changed: true, Online: online, At: t.since, Cause: t.cause}
}
