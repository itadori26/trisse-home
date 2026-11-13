package domain

import (
	"errors"
	"testing"
)

func widgetAt(id string, x, y, h int) Widget {
	return RestoreWidget(id, "t", "clock", Rect{X: x, Y: y, W: 4, H: h}, "{}")
}

func ids(m MobileLayout) []string {
	var out []string
	for _, e := range m.Entries() {
		out = append(out, e.WidgetID)
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDeriveMobileLayoutUsesReadingOrder(t *testing.T) {
	ws := []Widget{widgetAt("c", 0, 4, 2), widgetAt("b", 8, 0, 3), widgetAt("a", 0, 0, 5)}
	m := DeriveMobileLayout(ws)
	if !equal(ids(m), []string{"a", "b", "c"}) || m.Custom() {
		t.Fatalf("got %v custom=%v", ids(m), m.Custom())
	}
	if e := m.Entries()[0]; e.Rows != 5 || e.Hidden {
		t.Fatalf("desktop height expected, got %+v", e)
	}
}

func TestResolveMobileLayout(t *testing.T) {
	ws := []Widget{widgetAt("a", 0, 0, 2), widgetAt("b", 4, 0, 2), widgetAt("new", 0, 9, 3)}
	stored := []MobileEntry{
		{WidgetID: "b", Hidden: true, Rows: 40},
		{WidgetID: "gone", Rows: 2},
		{WidgetID: "a", Rows: 1},
		{WidgetID: "a", Rows: 7},
	}
	m := ResolveMobileLayout(stored, ws)
	if !m.Custom() || !equal(ids(m), []string{"b", "a", "new"}) {
		t.Fatalf("got %v", ids(m))
	}
	e := m.Entries()
	if !e[0].Hidden || e[0].Rows != MaxWidgetRows || e[1].Rows != 1 || e[2].Rows != 3 {
		t.Fatalf("entries %+v", e)
	}
	if ResolveMobileLayout(nil, ws).Custom() {
		t.Fatal("nil must derive")
	}
	if !ResolveMobileLayout([]MobileEntry{}, ws).Custom() {
		t.Fatal("an empty stored layout is still custom")
	}
}

func TestNewMobileLayoutValidates(t *testing.T) {
	ws := []Widget{widgetAt("a", 0, 0, 2), widgetAt("b", 4, 0, 2)}
	if _, err := NewMobileLayout([]MobileEntry{{WidgetID: "x", Rows: 2}}, ws); !errors.Is(err, ErrWidgetNotFound) {
		t.Fatalf("foreign widget: %v", err)
	}
	for _, bad := range [][]MobileEntry{
		{{WidgetID: "a", Rows: 2}, {WidgetID: "a", Rows: 2}},
		{{WidgetID: "a", Rows: 0}},
		{{WidgetID: "a", Rows: 25}},
	} {
		if _, err := NewMobileLayout(bad, ws); err == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
	got, err := NewMobileLayout([]MobileEntry{{WidgetID: "b", Rows: 3, Hidden: true}}, ws)
	if err != nil || len(got) != 1 {
		t.Fatalf("%v %v", got, err)
	}
}
