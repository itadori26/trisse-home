package domain

import (
	"net/netip"
	"testing"
	"time"
)

func rules(t *testing.T, in ...RuleInput) []TabRule {
	t.Helper()
	r, err := NewTabRules(in)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Thursday 12 November 2026.
func at(h, m int) time.Time { return time.Date(2026, 11, 12, h, m, 0, 0, time.UTC) }

func TestTabRuleMatches(t *testing.T) {
	lan := netip.MustParseAddr("192.168.178.20")
	vpn := netip.MustParseAddr("10.13.13.2")
	cases := []struct {
		name string
		rule RuleInput
		now  time.Time
		ip   netip.Addr
		want bool
	}{
		{"morning range", RuleInput{From: "06:00", To: "10:00"}, at(7, 30), lan, true},
		{"before range", RuleInput{From: "06:00", To: "10:00"}, at(5, 59), lan, false},
		{"end is exclusive", RuleInput{From: "06:00", To: "10:00"}, at(10, 0), lan, false},
		{"weekday", RuleInput{Days: []int{4}}, at(12, 0), lan, true},
		{"other weekday", RuleInput{Days: []int{1, 2}}, at(12, 0), lan, false},
		{"over midnight, late", RuleInput{Days: []int{4}, From: "22:00", To: "02:00"}, at(23, 0), lan, true},
		{"over midnight, early hours belong to the day before", RuleInput{Days: []int{3}, From: "22:00", To: "02:00"}, at(1, 0), lan, true},
		{"over midnight, early hours of the same weekday do not count", RuleInput{Days: []int{4}, From: "22:00", To: "02:00"}, at(1, 0), lan, false},
		{"over midnight, outside", RuleInput{From: "22:00", To: "02:00"}, at(12, 0), lan, false},
		{"network", RuleInput{Networks: []string{"192.168.178.0/24"}}, at(12, 0), lan, true},
		{"other network", RuleInput{Networks: []string{"192.168.178.0/24"}}, at(12, 0), vpn, false},
		{"IPv4-mapped IPv6", RuleInput{Networks: []string{"192.168.178.0/24"}}, at(12, 0), netip.MustParseAddr("::ffff:192.168.178.20"), true},
		{"unknown client", RuleInput{Networks: []string{"10.0.0.0/8"}}, at(12, 0), netip.Addr{}, false},
		{"everything", RuleInput{}, at(3, 3), netip.Addr{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rules(t, tc.rule)[0].Matches(tc.now, tc.ip); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
}

func TestNewTabRulesValidates(t *testing.T) {
	for _, bad := range []RuleInput{
		{Days: []int{7}},
		{From: "25:00", To: "26:00"},
		{From: "07:00"},
		{Networks: []string{"lan"}},
	} {
		if _, err := NewTabRules([]RuleInput{bad}); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	r := rules(t, RuleInput{Days: []int{1, 5}, From: "07:30", To: "09:00", Networks: []string{"192.168.178.7/24"}})[0]
	if f, to := r.Range(); f != "07:30" || to != "09:00" || r.Networks()[0] != "192.168.178.0/24" || len(r.Days()) != 2 {
		t.Fatalf("%v %v %v %v", f, to, r.Networks(), r.Days())
	}
	from, to := r.Minutes()
	if RestoreTabRule(r.Days(), from, to, r.Networks()).Matches(time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC), netip.MustParseAddr("192.168.178.9")) != true {
		t.Fatal("restored rule")
	}
}

func TestDefaultTabID(t *testing.T) {
	home := RestoreTab("home", "Home", 0, "").WithRules(rules(t, RuleInput{From: "06:00", To: "10:00"}))
	evening := RestoreTab("eve", "Evening", 1, "").WithRules(rules(t, RuleInput{From: "18:00", To: "23:00"}, RuleInput{From: "07:00", To: "08:00"}))
	plain := RestoreTab("plain", "Plain", 2, "")
	tabs := []Tab{home, evening, plain}
	if got := DefaultTabID(tabs, at(7, 30), netip.Addr{}); got != "home" {
		t.Fatalf("conflict must be resolved by order, got %q", got)
	}
	if got := DefaultTabID(tabs, at(19, 0), netip.Addr{}); got != "eve" {
		t.Fatalf("got %q", got)
	}
	if got := DefaultTabID(tabs, at(12, 0), netip.Addr{}); got != "" {
		t.Fatalf("no rule must keep the last tab, got %q", got)
	}
}
