package domain

import (
	"testing"
	"time"
)

func at(day time.Weekday, h, m int) time.Time {
	// 2026-11-09 is a Monday.
	return time.Date(2026, 11, 8+int(day), h, m, 0, 0, time.UTC)
}

func TestQuietHours(t *testing.T) {
	night := QuietHours{Enabled: true, Start: "22:00", End: "07:00", Days: []time.Weekday{time.Friday}}
	cases := []struct {
		t    time.Time
		want bool
	}{
		{at(time.Friday, 23, 0), true},
		{at(time.Saturday, 6, 59), true}, // morning part belongs to Friday
		{at(time.Saturday, 7, 0), false},
		{at(time.Thursday, 23, 0), false},
		{at(time.Friday, 6, 0), false}, // belongs to Thursday
	}
	for _, c := range cases {
		if got := night.Active(c.t); got != c.want {
			t.Errorf("%s: %v, want %v", c.t.Format("Mon 15:04"), got, c.want)
		}
	}
	day := QuietHours{Enabled: true, Start: "12:00", End: "13:00"}
	if !day.Active(at(time.Tuesday, 12, 30)) || day.Active(at(time.Tuesday, 13, 0)) {
		t.Fatal("daytime quiet hours wrong")
	}
	if (QuietHours{Enabled: true, Start: "25:00", End: "07:00"}).Validate() == nil || (QuietHours{Enabled: true, Start: "7:00", End: "08:00"}).Validate() == nil {
		t.Fatal("invalid times accepted")
	}
}

func TestPolicyDecide(t *testing.T) {
	p := Policy{Severities: map[string]Severity{"container": SeverityInfo, "x": "bogus"}, Quiet: QuietHours{Enabled: true, Start: "22:00", End: "07:00"}}
	if p.SeverityOf("container") != SeverityInfo || p.SeverityOf("service") != SeverityCritical || p.SeverityOf("x") != SeverityWarning {
		t.Fatal("severities wrong")
	}
	night := at(time.Monday, 23, 0)
	if p.Decide(SeverityCritical, night) != Deliver || p.Decide(SeverityWarning, night) != Defer || p.Decide(SeverityWarning, at(time.Monday, 12, 0)) != Deliver {
		t.Fatal("decisions wrong")
	}
}

func TestMaintenanceWindow(t *testing.T) {
	now := at(time.Monday, 10, 0)
	w, err := NewMaintenanceWindow("1", []string{"service:NAS", " ", "container:nginx"}, now, now.Add(time.Hour), "update")
	if err != nil {
		t.Fatal(err)
	}
	if !w.ActiveAt(now) || w.ActiveAt(now.Add(time.Hour)) || !w.Covers("service", "id-1", "nas") || !w.Covers("container", "nginx", "") || w.Covers("service", "x", "web") {
		t.Fatal("coverage wrong")
	}
	for _, scope := range [][]string{nil, {"disk:/"}, {"nas"}} {
		if _, err := NewMaintenanceWindow("1", scope, now, now.Add(time.Hour), ""); err == nil {
			t.Errorf("scope %v accepted", scope)
		}
	}
	if _, err := NewMaintenanceWindow("1", []string{"*"}, now, now.Add(8*24*time.Hour), ""); err == nil {
		t.Fatal("too long accepted")
	}
}

func TestChannel(t *testing.T) {
	c := Channel{Kind: ChannelWebhook, Name: " Hook ", URL: "https://example.org/hook", Enabled: true, MinSeverity: SeverityWarning, Types: []string{"service"}}
	if err := c.Validate(); err != nil || c.Template != DefaultWebhookTemplate || c.Name != "Hook" {
		t.Fatalf("%+v %v", c, err)
	}
	if c.Accepts(Message{Type: "service", Severity: SeverityInfo}) || !c.Accepts(Message{Type: "service", Severity: SeverityCritical}) ||
		c.Accepts(Message{Type: "disk", Severity: SeverityCritical}) || !c.Accepts(Message{Type: "digest", Severity: SeverityWarning}) {
		t.Fatal("routing wrong")
	}
	mail := Channel{Kind: ChannelEmail, Name: "Mail", SMTPHost: "smtp.example.org", From: "home@example.org", To: "me@example.org, you@example.org"}
	if err := mail.Validate(); err != nil || mail.SMTPPort != 587 || mail.TLS != "starttls" {
		t.Fatalf("%+v %v", mail, err)
	}
	for _, bad := range []Channel{
		{Kind: ChannelGotify, Name: "x", URL: "ftp://x"},
		{Kind: ChannelEmail, Name: "x", SMTPHost: "smtp", From: "nope", To: "me@example.org"},
		{Kind: "sms", Name: "x"},
		{Kind: ChannelWebhook, Name: "x", URL: "https://x", Types: []string{"weather"}},
	} {
		if bad.Validate() == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
}

func TestAlertKinds(t *testing.T) {
	cases := []struct {
		kind     string
		tag      string
		severity Severity
	}{
		{"security", "lock", SeverityWarning},
		{"reminder", "bell", SeverityInfo},
		{"briefing", "sunny", SeverityInfo},
		{"unknown", "lock", SeverityWarning},
	}
	for _, c := range cases {
		if got := AlertTag(c.kind); got != c.tag {
			t.Errorf("AlertTag(%q) = %q, want %q", c.kind, got, c.tag)
		}
		if got := (Policy{}).SeverityOf(c.kind); got != c.severity {
			t.Errorf("SeverityOf(%q) = %q, want %q", c.kind, got, c.severity)
		}
	}
	reminder := Channel{Name: "Phone", Kind: ChannelWebhook, URL: "https://example.org/hook", Types: []string{"reminder", "briefing"}}
	if err := reminder.Validate(); err != nil {
		t.Fatalf("reminder and briefing must be selectable per channel: %v", err)
	}
}
