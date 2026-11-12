package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	s     Settings
	ok    bool
	saves int
}

func (m *memRepo) Load(context.Context) (Settings, bool, error) { return m.s, m.ok, nil }
func (m *memRepo) Save(_ context.Context, s Settings) error {
	m.s, m.ok = s, true
	m.saves++
	return nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type todos []TodoFact

func (f todos) DueTodos(_ context.Context, until time.Time) ([]TodoFact, error) {
	var out []TodoFact
	for _, t := range f {
		if !t.Due.After(until) {
			out = append(out, t)
		}
	}
	return out, nil
}

type events []EventFact

func (f events) UpcomingEvents(context.Context, time.Time) ([]EventFact, error) { return f, nil }

type failingWarnings struct{}

func (failingWarnings) WeatherWarnings(context.Context, float64, float64) ([]WarningFact, error) {
	return nil, errors.New("DWD down")
}

type warnings struct{ lat, lon float64 }

func (w *warnings) WeatherWarnings(_ context.Context, lat, lon float64) ([]WarningFact, error) {
	w.lat, w.lon = lat, lon
	return []WarningFact{{Headline: "Storm", Level: 3}}, nil
}

type collections []CollectionFact

func (f collections) Collections(context.Context, time.Time) ([]CollectionFact, error) { return f, nil }

type services []ServiceFact

func (f services) DownServices(context.Context) ([]ServiceFact, error) { return f, nil }

type incidents struct{ since time.Time }

func (i *incidents) Incidents(_ context.Context, since time.Time) ([]IncidentFact, error) {
	i.since = since
	return []IncidentFact{{Message: "Backup failed", Critical: true, At: since.Add(time.Hour)}}, nil
}

type publisher struct {
	titles, bodies []string
	urgent         bool
}

func (p *publisher) Publish(_ context.Context, title, body string, urgent bool) {
	p.titles, p.bodies, p.urgent = append(p.titles, title), append(p.bodies, body), urgent
}

func sectionKinds(v BriefingView) []domain.SectionKind {
	var out []domain.SectionKind
	for _, s := range v.Briefing.Sections {
		out = append(out, s.Kind)
	}
	return out
}

var cet = time.FixedZone("CET", 3600)

func TestBuildBriefing(t *testing.T) {
	ctx := context.Background()
	// Thursday 12 November 2026, 07:30 local.
	c := &clk{t: time.Date(2026, 11, 12, 6, 30, 0, 0, time.UTC)}
	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	inc, warn := &incidents{}, &warnings{}
	d := Deps{Repo: &memRepo{}, Clock: c, Location: cet, DefaultLat: 52.5, DefaultLon: 13.4, Sources: Sources{
		Todos:      todos{{Title: "Rent", Due: day(10), Overdue: true}, {Title: "Plants", Due: day(12)}, {Title: "Later", Due: day(20)}},
		Events:     events{{Title: "Dentist", Start: time.Date(2026, 11, 12, 9, 0, 0, 0, cet)}},
		Warnings:   warn,
		Collection: collections{{Type: "Paper", Day: day(13)}},
		Services:   services{{Name: "NAS", Error: "timeout"}},
		Incidents:  inc,
	}}
	v, err := NewBuildBriefing(d).Execute(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.SectionKind{domain.SectionWarnings, domain.SectionIncidents, domain.SectionServices, domain.SectionTodos, domain.SectionEvents, domain.SectionWaste}
	if got := sectionKinds(v); len(got) != len(want) || got[0] != want[0] || got[5] != want[5] {
		t.Fatalf("sections %v", got)
	}
	if warn.lat != 52.5 || warn.lon != 13.4 {
		t.Fatal("default location not used")
	}
	if !inc.since.Equal(time.Date(2026, 11, 11, 18, 0, 0, 0, cet)) {
		t.Fatalf("incidents since %v", inc.since)
	}
	todosSec := v.Briefing.Sections[3]
	if len(todosSec.Items) != 2 || todosSec.Items[0].Text != "Rent" || todosSec.Items[0].Detail != "overdue" {
		t.Fatalf("todos %+v", todosSec.Items)
	}
	if w := v.Briefing.Sections[5].Items[0]; w.Detail != "tomorrow" {
		t.Fatalf("waste %+v", w)
	}
	if v.Briefing.Sections[0].Items[0].Severity != domain.SeverityCritical {
		t.Fatal("level 3 warnings are critical")
	}

	// A failing source hides its section and is reported; missing sources
	// (disabled modules) are simply absent.
	d.Sources.Warnings = failingWarnings{}
	d.Sources.Collection = nil
	v, _ = NewBuildBriefing(d).Execute(ctx)
	if len(v.Unavailable) != 1 || v.Unavailable[0] != domain.SectionWarnings || len(v.Briefing.Sections) != 4 {
		t.Fatalf("unavailable %v sections %v", v.Unavailable, sectionKinds(v))
	}
	s, _ := NewManageSettings(d).Get(ctx)
	if len(s.Available) != 5 || s.PushOn || s.PushAt != "07:00" {
		t.Fatalf("settings %+v", s)
	}
}

func TestSettingsAndPush(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 12, 5, 59, 0, 0, time.UTC)} // 06:59 local, Thursday
	repo, pub := &memRepo{}, &publisher{}
	d := Deps{Repo: repo, Clock: c, Location: cet, Publisher: pub, Sources: Sources{
		Services: services{{Name: "NAS"}},
	}}
	settings, send := NewManageSettings(d), NewSendBriefing(d)

	if sent, _ := send.Execute(ctx); sent {
		t.Fatal("off by default")
	}
	if _, err := settings.Save(ctx, SettingsInput{PushOn: true, PushAt: "7:00"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid time: %v", err)
	}
	if _, err := settings.Save(ctx, SettingsInput{PushOn: true, PushAt: "07:00", Lat: 91}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid location: %v", err)
	}
	v, err := settings.Save(ctx, SettingsInput{PushOn: true, PushAt: "07:00", PushDays: []int{4}, SkipEmpty: true,
		Layout: []domain.LayoutEntry{{Kind: domain.SectionServices, Enabled: true}}})
	if err != nil || !v.PushOn || len(v.Layout) != len(domain.DefaultOrder) {
		t.Fatalf("save %+v %v", v, err)
	}

	if sent, _ := send.Execute(ctx); sent {
		t.Fatal("not yet 07:00")
	}
	c.t = c.t.Add(time.Minute)
	if sent, err := send.Execute(ctx); !sent || err != nil || len(pub.titles) != 1 || !pub.urgent ||
		!strings.Contains(pub.bodies[0], "Offline:\n• NAS") || pub.titles[0] != "Briefing for Thu, 12 Nov" {
		t.Fatalf("push %v %v %+v", sent, err, pub)
	}
	c.t = c.t.Add(time.Hour)
	if sent, _ := send.Execute(ctx); sent {
		t.Fatal("once a day")
	}
	// Saving settings keeps the time of the last push.
	if _, err := settings.Save(ctx, SettingsInput{PushOn: true, PushAt: "07:00", PushDays: []int{4, 5}, SkipEmpty: true}); err != nil {
		t.Fatal(err)
	}
	if sent, _ := send.Execute(ctx); sent {
		t.Fatal("saving settings re-armed the push")
	}

	// Friday: nothing to report and SkipEmpty → nothing sent, but the day
	// counts as done.
	d.Sources.Services = services{}
	send = NewSendBriefing(d)
	c.t = time.Date(2026, 11, 13, 6, 0, 0, 0, time.UTC)
	if sent, _ := send.Execute(ctx); sent || len(pub.titles) != 1 {
		t.Fatal("empty briefing must be skipped")
	}
	if !repo.s.LastSent.Equal(c.t.In(cet)) {
		t.Fatalf("last sent %v", repo.s.LastSent)
	}
	// The test button sends regardless of schedule and content.
	if err := send.SendNow(ctx); err != nil || len(pub.titles) != 2 || !strings.Contains(pub.bodies[1], "Nothing to report") {
		t.Fatalf("send now %v %+v", err, pub)
	}
	d.Publisher = nil
	if err := NewSendBriefing(d).SendNow(ctx); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("without notifications: %v", err)
	}
}
