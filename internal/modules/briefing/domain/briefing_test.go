package domain

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func kinds(b Briefing) []SectionKind {
	var out []SectionKind
	for _, s := range b.Sections {
		out = append(out, s.Kind)
	}
	return out
}

func TestNormaliseLayout(t *testing.T) {
	got := NormaliseLayout([]LayoutEntry{{Kind: SectionTodos}, {Kind: "unknown", Enabled: true}, {Kind: SectionTodos, Enabled: true}, {Kind: SectionWaste, Enabled: true}})
	if len(got) != len(DefaultOrder) || got[0] != (LayoutEntry{Kind: SectionTodos}) || got[1].Kind != SectionWaste || !got[2].Enabled || got[2].Kind != SectionWarnings {
		t.Fatalf("got %+v", got)
	}
}

func TestCompose(t *testing.T) {
	now := time.Date(2026, 11, 12, 7, 0, 0, 0, time.UTC)
	items := map[SectionKind][]Item{
		SectionTodos: {
			{Text: "later", At: now.Add(2 * time.Hour)},
			{Text: "overdue", At: now.Add(-48 * time.Hour), Severity: SeverityWarning},
			{Text: "soon", At: now.Add(time.Hour)},
		},
		SectionEvents:   {{Text: "Dentist", At: now.Add(3 * time.Hour)}},
		SectionWaste:    {},
		SectionServices: {{Text: "NAS"}, {Text: "a"}, {Text: "b"}, {Text: "c"}, {Text: "d"}, {Text: "e"}, {Text: "f"}},
	}
	layout := Layout{{Kind: SectionEvents, Enabled: true}, {Kind: SectionServices, Enabled: false}}
	b := Compose(layout, items)
	if !reflect.DeepEqual(kinds(b), []SectionKind{SectionEvents, SectionTodos}) {
		t.Fatalf("sections %v", kinds(b))
	}
	if todos := b.Sections[1].Items; todos[0].Text != "overdue" || todos[1].Text != "soon" || todos[2].Text != "later" {
		t.Fatalf("todo order %+v", todos)
	}
	full := Compose(DefaultLayout(), items)
	if s := full.Sections[0]; s.Kind != SectionServices || len(s.Items) != MaxItemsPerSection || s.More != 2 {
		t.Fatalf("cap %+v", s)
	}
	if !Compose(DefaultLayout(), nil).Empty() {
		t.Fatal("no items, no sections")
	}
}

func TestWindows(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	now := time.Date(2026, 11, 12, 7, 30, 0, 0, loc)
	if got := IncidentsSince(now); !got.Equal(time.Date(2026, 11, 11, 18, 0, 0, 0, loc)) {
		t.Fatalf("since %v", got)
	}
	if got := EndOfTomorrow(now); !got.Equal(time.Date(2026, 11, 14, 0, 0, 0, 0, loc)) {
		t.Fatalf("until %v", got)
	}
}

func TestPushSchedule(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	s, err := NewPushSchedule(true, "07:00", []int{4, 1, 4}, true)
	if err != nil || s.At() != "07:00" || !reflect.DeepEqual(s.DayNumbers(), []int{1, 4}) {
		t.Fatalf("%+v %v", s, err)
	}
	for _, bad := range []struct {
		at   string
		days []int
	}{{"7:00", nil}, {"25:00", nil}, {"07:00", []int{7}}} {
		if _, err := NewPushSchedule(true, bad.at, bad.days, false); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	thu := func(h, m int) time.Time { return time.Date(2026, 11, 12, h, m, 0, 0, loc) } // Thursday
	never := time.Time{}
	if s.Due(thu(6, 59), never) || !s.Due(thu(7, 0), never) || !s.Due(thu(12, 0), never) {
		t.Fatal("fires at 07:00 and catches up later that day")
	}
	if s.Due(thu(8, 0), thu(7, 0)) {
		t.Fatal("sent once a day")
	}
	if !s.Due(thu(7, 0), thu(7, 0).AddDate(0, 0, -3)) {
		t.Fatal("sent again on the next enabled day")
	}
	if s.Due(time.Date(2026, 11, 13, 9, 0, 0, 0, loc), never) {
		t.Fatal("Friday is not enabled")
	}
	s.Enabled = false
	if s.Due(thu(9, 0), never) {
		t.Fatal("disabled")
	}
	if d := DefaultPushSchedule(); d.Enabled || d.At() != "07:00" || !d.SkipEmpty {
		t.Fatalf("default %+v", d)
	}
}

func TestText(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	b := Briefing{Sections: []Section{
		{Kind: SectionWarnings, Items: []Item{{Text: "Storm", Detail: "until 18:00", Severity: SeverityCritical}}},
		{Kind: SectionEvents, Items: []Item{{Text: "Dentist", At: time.Date(2026, 11, 12, 8, 30, 0, 0, time.UTC)}}, More: 2},
	}}
	want := "Weather warnings:\n• Storm – until 18:00\n\nAppointments:\n• Dentist (09:30)\n• … and 2 more"
	if got := b.Text(loc); got != want {
		t.Fatalf("got %q", got)
	}
	if b.MostSevere() != SeverityCritical || (Briefing{}).Text(loc) != "Nothing to report." {
		t.Fatal("severity / empty text")
	}
	if !strings.Contains(b.Text(loc), "Storm") {
		t.Fatal("text")
	}
}
