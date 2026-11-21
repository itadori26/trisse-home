package domain

import "time"

// Health is the traffic light of the backups (#677).
type Health string

// The values are part of the API of GET /api/backups/status.
const (
	// HealthOK means the newest backup is as recent as the interval promises.
	HealthOK Health = "ok"
	// HealthStale means the newest backup is older than expected, there is
	// none yet, or the off-site copy failed while the local backup worked.
	HealthStale Health = "stale"
	// HealthFailed means the latest backup attempt failed.
	HealthFailed Health = "failed"
)

// minGrace is the least slack before a backup counts as overdue: the job
// checks once an hour, so a backup can be up to an hour late on schedule.
const minGrace = 2 * time.Hour

// Outcome is the result of the latest backup attempt and off-site upload;
// zero times mean there was none since the start.
type Outcome struct {
	FailedAt time.Time
	Error    string
	// OffsiteFailedAt is set while the newest upload failed.
	OffsiteFailedAt time.Time
	OffsiteError    string
}

// StaleAfter is how old the newest backup may get before it counts as
// overdue: one interval plus half an interval of slack, at least minGrace.
func (p Policy) StaleAfter() time.Duration {
	return p.Interval + max(p.Interval/2, minGrace)
}

// Health judges the backups from the newest file and the latest outcome. A
// failure counts only while no backup has been taken after it.
func (p Policy) Health(latest time.Time, o Outcome, now time.Time) Health {
	switch {
	case !o.FailedAt.IsZero() && o.FailedAt.After(latest):
		return HealthFailed
	case latest.IsZero(), now.Sub(latest) > p.StaleAfter(), !o.OffsiteFailedAt.IsZero():
		return HealthStale
	}
	return HealthOK
}
