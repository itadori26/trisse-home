package domain

import (
	"testing"
	"time"
)

func TestCronPrev(t *testing.T) {
	at := func(s string) time.Time {
		v, _ := time.Parse("2006-01-02 15:04", s)
		return v
	}
	cases := []struct{ expr, now, want string }{
		{"0 3 * * *", "2026-11-12 10:00", "2026-11-12 03:00"},
		{"0 3 * * *", "2026-11-12 02:59", "2026-11-11 03:00"},
		{"*/15 * * * *", "2026-11-12 10:44", "2026-11-12 10:30"},
		{"30 2 * * 1-5", "2026-11-16 01:00", "2026-11-13 02:30"}, // Monday 01:00 -> Friday
		{"0 0 1 * *", "2026-11-12 10:00", "2026-11-01 00:00"},
		{"0 12 * * 7", "2026-11-12 10:00", "2026-11-08 12:00"}, // 7 = Sunday
		{"5/20 8 * * *", "2026-11-12 08:30", "2026-11-12 08:25"},
		{"@daily", "2026-11-12 10:00", "2026-11-12 00:00"},
		{"0 9 13 * 1", "2026-11-12 10:00", "2026-11-09 09:00"}, // day 13 OR Monday
	}
	for _, c := range cases {
		cr, err := ParseCron(c.expr)
		if err != nil {
			t.Fatalf("%s: %v", c.expr, err)
		}
		if got := cr.Prev(at(c.now)); !got.Equal(at(c.want)) {
			t.Errorf("%s at %s: %s, want %s", c.expr, c.now, got.Format("2006-01-02 15:04"), c.want)
		}
	}
	for _, bad := range []string{"", "* * * *", "60 * * * *", "* 24 * * *", "*/0 * * * *", "a b c d e", "5-1 * * * *"} {
		if _, err := ParseCron(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
