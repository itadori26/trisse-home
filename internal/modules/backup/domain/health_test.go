package domain

import (
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	daily, _ := NewPolicy(24*time.Hour, 7)
	hourly, _ := NewPolicy(time.Hour, 7)
	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) time.Time { return now.Add(-d) }

	tests := []struct {
		name   string
		policy Policy
		latest time.Time
		o      Outcome
		want   Health
	}{
		{"fresh", daily, ago(9 * time.Hour), Outcome{}, HealthOK},
		{"late but within the slack", daily, ago(35 * time.Hour), Outcome{}, HealthOK},
		{"overdue", daily, ago(37 * time.Hour), Outcome{}, HealthStale},
		{"hourly uses the minimum grace", hourly, ago(150 * time.Minute), Outcome{}, HealthOK},
		{"hourly overdue", hourly, ago(4 * time.Hour), Outcome{}, HealthStale},
		{"none yet", daily, time.Time{}, Outcome{}, HealthStale},
		{"failed after the newest backup", daily, ago(9 * time.Hour), Outcome{FailedAt: ago(time.Hour), Error: "disk full"}, HealthFailed},
		{"failed without any backup", daily, time.Time{}, Outcome{FailedAt: ago(time.Hour)}, HealthFailed},
		{"failure followed by a backup", daily, ago(time.Hour), Outcome{FailedAt: ago(2 * time.Hour)}, HealthOK},
		{"off-site upload failed", daily, ago(time.Hour), Outcome{OffsiteFailedAt: ago(time.Hour)}, HealthStale},
	}
	for _, tt := range tests {
		if got := tt.policy.Health(tt.latest, tt.o, now); got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.name, got, tt.want)
		}
	}
}
