package domain

import (
	"fmt"
	"net/netip"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Limits of tab rules (#71).
const (
	MaxTabRules     = 10
	MaxRuleNetworks = 10
	minutesPerDay   = 24 * 60
	ruleTimeLayout  = "15:04"
)

// TabRule opens a tab by default on certain weekdays, in a time range
// and/or from certain networks (#71). Empty parts match always.
type TabRule struct {
	days     []time.Weekday
	from, to int // minutes since midnight; from == to means the whole day
	networks []netip.Prefix
}

// RuleInput is a rule as entered by the user.
type RuleInput struct {
	Days     []int
	From, To string // "HH:MM" or "" for the whole day
	Networks []string
}

func parseMinutes(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	t, err := time.Parse(ruleTimeLayout, s)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// NewTabRules validates rules.
func NewTabRules(in []RuleInput) ([]TabRule, error) {
	if len(in) > MaxTabRules {
		return nil, kernel.NewValidationError("rules", "At most 10 rules per tab are supported.")
	}
	out := make([]TabRule, 0, len(in))
	invalid := kernel.NewValidationError("rules", "Rules need weekdays 0–6, times like 07:30 and networks like 192.168.178.0/24.")
	for _, r := range in {
		var rule TabRule
		for _, d := range r.Days {
			if d < 0 || d > 6 {
				return nil, invalid
			}
			rule.days = append(rule.days, time.Weekday(d))
		}
		var err1, err2 error
		rule.from, err1 = parseMinutes(strings.TrimSpace(r.From))
		rule.to, err2 = parseMinutes(strings.TrimSpace(r.To))
		if err1 != nil || err2 != nil || (r.From == "") != (r.To == "") {
			return nil, invalid
		}
		if len(r.Networks) > MaxRuleNetworks {
			return nil, invalid
		}
		for _, n := range r.Networks {
			p, err := netip.ParsePrefix(strings.TrimSpace(n))
			if err != nil {
				return nil, invalid
			}
			rule.networks = append(rule.networks, p.Masked())
		}
		out = append(out, rule)
	}
	return out, nil
}

// RestoreTabRule rebuilds a stored rule; invalid parts are dropped.
func RestoreTabRule(days []int, from, to int, networks []string) TabRule {
	r := TabRule{from: from % minutesPerDay, to: to % minutesPerDay}
	for _, d := range days {
		if d >= 0 && d <= 6 {
			r.days = append(r.days, time.Weekday(d))
		}
	}
	for _, n := range networks {
		if p, err := netip.ParsePrefix(n); err == nil {
			r.networks = append(r.networks, p.Masked())
		}
	}
	return r
}

// Days returns the weekdays (empty = every day).
func (r TabRule) Days() []int {
	out := make([]int, 0, len(r.days))
	for _, d := range r.days {
		out = append(out, int(d))
	}
	return out
}

// Minutes returns the time range in minutes since midnight.
func (r TabRule) Minutes() (from, to int) { return r.from, r.to }

// Range returns the time range as "HH:MM" ("" for the whole day).
func (r TabRule) Range() (from, to string) {
	if r.from == r.to {
		return "", ""
	}
	f := func(m int) string { return fmt.Sprintf("%02d:%02d", m/60, m%60) }
	return f(r.from), f(r.to)
}

// Networks returns the CIDR ranges (empty = any network).
func (r TabRule) Networks() []string {
	out := make([]string, 0, len(r.networks))
	for _, p := range r.networks {
		out = append(out, p.String())
	}
	return out
}

// Matches reports whether the rule applies at now for a client at ip.
// A range over midnight (22:00–02:00) belongs to the weekday it starts on.
func (r TabRule) Matches(now time.Time, ip netip.Addr) bool {
	if len(r.networks) > 0 {
		ip = ip.Unmap()
		found := false
		for _, p := range r.networks {
			if ip.IsValid() && p.Contains(ip) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	m := now.Hour()*60 + now.Minute()
	day := now.Weekday()
	switch {
	case r.from == r.to: // whole day
	case r.from < r.to:
		if m < r.from || m >= r.to {
			return false
		}
	default: // over midnight
		if m < r.to {
			day = (day + 6) % 7 // the early hours belong to the previous day's rule
		} else if m < r.from {
			return false
		}
	}
	if len(r.days) == 0 {
		return true
	}
	for _, d := range r.days {
		if d == day {
			return true
		}
	}
	return false
}

// DefaultTabID returns the first tab (in tab order) with a matching rule,
// or "" so the client keeps its last used tab.
func DefaultTabID(tabs []Tab, now time.Time, ip netip.Addr) string {
	for _, t := range tabs {
		for _, r := range t.rules {
			if r.Matches(now, ip) {
				return t.id
			}
		}
	}
	return ""
}
