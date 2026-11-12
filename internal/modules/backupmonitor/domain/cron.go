package domain

import (
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Cron is a parsed five-field cron expression (minute hour day-of-month
// month day-of-week) as used by crontab (#100).
type Cron struct {
	expr                          string
	minute, hour, dom, month, dow [64]bool
	domAny, dowAny                bool
}

// ErrInvalidCron is returned for expressions that cannot be parsed.
var ErrInvalidCron = kernel.NewValidationError("cron", "Use a cron expression with five fields, e.g. 0 3 * * * (daily at 03:00).")

var cronAliases = map[string]string{"@hourly": "0 * * * *", "@daily": "0 0 * * *", "@midnight": "0 0 * * *", "@weekly": "0 0 * * 0", "@monthly": "0 0 1 * *"}

// ParseCron parses expressions such as "0 3 * * *", "*/15 * * * *",
// "30 2 * * 1-5" or "@daily".
func ParseCron(expr string) (Cron, error) {
	expr = strings.TrimSpace(expr)
	if alias, ok := cronAliases[strings.ToLower(expr)]; ok {
		expr = alias
	}
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return Cron{}, ErrInvalidCron
	}
	c := Cron{expr: strings.Join(fields, " ")}
	specs := []struct {
		set      *[64]bool
		min, max int
	}{{&c.minute, 0, 59}, {&c.hour, 0, 23}, {&c.dom, 1, 31}, {&c.month, 1, 12}, {&c.dow, 0, 7}}
	for i, f := range fields {
		if err := parseField(f, specs[i].set, specs[i].min, specs[i].max); err != nil {
			return Cron{}, err
		}
	}
	if c.dow[7] {
		c.dow[0] = true // 7 is Sunday as well
	}
	c.domAny, c.dowAny = fields[2] == "*", fields[4] == "*"
	return c, nil
}

func parseField(f string, set *[64]bool, minV, maxV int) error {
	for _, part := range strings.Split(f, ",") {
		step := 1
		if base, s, ok := strings.Cut(part, "/"); ok {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				return ErrInvalidCron
			}
			part, step = base, n
		}
		lo, hi := minV, maxV
		if part != "*" {
			a, b, isRange := strings.Cut(part, "-")
			var err1, err2 error
			lo, err1 = strconv.Atoi(a)
			hi = lo
			if isRange {
				hi, err2 = strconv.Atoi(b)
			}
			if err1 != nil || err2 != nil || lo < minV || hi > maxV || lo > hi {
				return ErrInvalidCron
			}
			if !isRange && step > 1 {
				hi = maxV // "5/15" means from 5 every 15
			}
		}
		for v := lo; v <= hi; v += step {
			set[v] = true
		}
	}
	return nil
}

// String returns the normalised expression.
func (c Cron) String() string { return c.expr }

// matchesDay applies cron's rule: if both day-of-month and day-of-week are
// restricted, either may match.
func (c Cron) matchesDay(t time.Time) bool {
	dom, dow := c.dom[t.Day()], c.dow[int(t.Weekday())]
	switch {
	case c.domAny && c.dowAny:
		return true
	case c.domAny:
		return dow
	case c.dowAny:
		return dom
	}
	return dom || dow
}

// Prev returns the latest scheduled minute at or before t (zero if none
// within a year).
func (c Cron) Prev(t time.Time) time.Time {
	t = t.Truncate(time.Minute)
	limit := t.AddDate(-1, 0, -1)
	for t.After(limit) {
		if !c.month[int(t.Month())] || !c.matchesDay(t) {
			// Jump to the last minute of the previous day.
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Add(-time.Minute)
			continue
		}
		if !c.hour[t.Hour()] {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location()).Add(-time.Minute)
			continue
		}
		if c.minute[t.Minute()] {
			return t
		}
		t = t.Add(-time.Minute)
	}
	return time.Time{}
}
