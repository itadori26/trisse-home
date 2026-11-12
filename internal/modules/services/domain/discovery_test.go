package domain

import (
	"testing"
	"time"
)

func TestParseLabels(t *testing.T) {
	cases := []struct {
		name   string
		c      Container
		want   DiscoveredService
		wantOK bool
	}{
		{"own labels", Container{Name: "jellyfin", Labels: map[string]string{
			"trisse-home.name": "Media", "trisse-home.url": "https://media.lan", "trisse-home.icon": "🎬", "trisse-home.description": "Films",
		}}, DiscoveredService{Container: "jellyfin", Name: "Media", URL: "https://media.lan", Icon: "🎬", Description: "Films"}, true},
		{"homepage fallback", Container{Name: "grafana", Labels: map[string]string{
			"homepage.name": "Grafana", "homepage.href": "http://grafana:3000", "homepage.icon": "grafana.png", "homepage.group": "Monitoring",
		}}, DiscoveredService{Container: "grafana", Name: "Grafana", URL: "http://grafana:3000", Icon: "grafana.png"}, true},
		{"own labels win over homepage", Container{Name: "x", Labels: map[string]string{
			"trisse-home.url": "https://a.lan", "homepage.href": "https://b.lan", "homepage.name": "B",
		}}, DiscoveredService{Container: "x", Name: "B", URL: "https://a.lan"}, true},
		{"name defaults to container", Container{Name: "pihole", Labels: map[string]string{"trisse-home.url": " https://pi.lan "}},
			DiscoveredService{Container: "pihole", Name: "pihole", URL: "https://pi.lan"}, true},
		{"no url", Container{Name: "db", Labels: map[string]string{"trisse-home.name": "DB"}}, DiscoveredService{}, false},
		{"opt out", Container{Name: "x", Labels: map[string]string{"trisse-home.url": "https://a.lan", "trisse-home.enable": "False"}}, DiscoveredService{}, false},
		{"no labels", Container{Name: "x"}, DiscoveredService{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := ParseLabels(c.c)
			if ok != c.wantOK || got != c.want {
				t.Fatalf("ParseLabels = %+v, %v; want %+v, %v", got, ok, c.want, c.wantOK)
			}
		})
	}
}

func TestDiscoveredServiceValidate(t *testing.T) {
	if err := (DiscoveredService{Name: "A", URL: "ftp://a"}).Validate(); err == nil {
		t.Fatal("a non-HTTP URL must be rejected")
	}
	if err := (DiscoveredService{Name: "A", URL: "https://a.lan"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPlanSync(t *testing.T) {
	manual := Service{ID: "m", Name: "Router", URL: "http://fritz.box", CheckURL: "http://fritz.box"}
	media := Service{ID: "1", Name: "Media", URL: "https://media.lan", CheckURL: "https://media.lan", Container: "jellyfin"}
	edited := Service{ID: "2", Name: "My Grafana", URL: "https://g.lan", CheckURL: "https://g.lan/health", Container: "grafana", Edited: true}
	gone := Service{ID: "3", Name: "Old", URL: "https://old.lan", CheckURL: "https://old.lan", Container: "old"}
	back := Service{ID: "4", Name: "Back", URL: "https://back.lan", CheckURL: "https://back.lan", Container: "back", Hidden: true}
	existing := []Service{manual, media, edited, gone, back}
	found := []DiscoveredService{
		{Container: "jellyfin", Name: "Jellyfin", URL: "https://jf.lan"},
		{Container: "grafana", Name: "Grafana", URL: "https://grafana.lan"},
		{Container: "back", Name: "Back", URL: "https://back.lan"},
		{Container: "new", Name: "New", URL: "https://new.lan"},
		{Container: "new", Name: "Duplicate", URL: "https://dup.lan"},
	}

	t.Run("labels win", func(t *testing.T) {
		p := PlanSync(existing, found, LabelsWin)
		if len(p.Create) != 1 || p.Create[0].Name != "New" {
			t.Fatalf("create: %+v", p.Create)
		}
		if len(p.Hide) != 1 || p.Hide[0].ID != "3" || !p.Hide[0].Hidden {
			t.Fatalf("hide: %+v", p.Hide)
		}
		if len(p.Show) != 1 || p.Show[0].ID != "4" || p.Show[0].Hidden {
			t.Fatalf("show: %+v", p.Show)
		}
		if len(p.Update) != 2 {
			t.Fatalf("update: %+v", p.Update)
		}
		jf, gr := p.Update[0], p.Update[1]
		if jf.Name != "Jellyfin" || jf.URL != "https://jf.lan" || jf.CheckURL != "https://jf.lan" {
			t.Fatalf("the check URL follows the URL: %+v", jf)
		}
		if gr.Name != "Grafana" || gr.CheckURL != "https://g.lan/health" || gr.Edited {
			t.Fatalf("labels must overwrite edits but keep a separate check URL: %+v", gr)
		}
	})

	t.Run("manual wins", func(t *testing.T) {
		p := PlanSync(existing, found, ManualWins)
		if len(p.Update) != 1 || p.Update[0].ID != "1" {
			t.Fatalf("edited tile must stay untouched: %+v", p.Update)
		}
	})

	t.Run("unchanged", func(t *testing.T) {
		p := PlanSync([]Service{media}, []DiscoveredService{{Container: "jellyfin", Name: "Media", URL: "https://media.lan"}}, LabelsWin)
		if !p.Empty() {
			t.Fatalf("nothing changed: %+v", p)
		}
	})
}

func TestDiscoverySettings(t *testing.T) {
	s := DefaultDiscoverySettings()
	if s.Enabled || s.Validate() != nil {
		t.Fatalf("defaults must be valid and off: %+v", s)
	}
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	if s.Due(time.Time{}, now) {
		t.Fatal("disabled discovery is never due")
	}
	s.Enabled = true
	if !s.Due(time.Time{}, now) || s.Due(now.Add(-4*time.Minute), now) || !s.Due(now.Add(-5*time.Minute), now) {
		t.Fatal("due after the interval")
	}
	for _, bad := range []DiscoverySettings{{IntervalMinutes: 0, Rule: LabelsWin}, {IntervalMinutes: MaxDiscoveryInterval + 1, Rule: LabelsWin}, {IntervalMinutes: 5, Rule: "x"}} {
		if bad.Validate() == nil {
			t.Fatalf("invalid settings accepted: %+v", bad)
		}
	}
}
