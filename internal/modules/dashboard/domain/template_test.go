package domain

import (
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	w, _ := NewWidget("w1", "t1", "clock", Rect{X: 0, Y: 0, W: 3, H: 2}, `{"tz":"UTC"}`)
	tpl, err := NewTemplate("x", " Server ", []Widget{w})
	if err != nil || tpl.Name != "Server" || len(tpl.Widgets) != 1 || tpl.Widgets[0].Config != `{"tz":"UTC"}` {
		t.Fatalf("%+v %v", tpl, err)
	}
	if _, err := NewTemplate("x", strings.Repeat("a", 41), nil); err == nil {
		t.Fatal("long name accepted")
	}
	tab, _ := NewTab("t", "Home", 0)
	if err := tab.SetBackground("javascript:alert(1)"); err == nil {
		t.Fatal("non-http background accepted")
	}
	if err := tab.SetBackground(" https://img.example/bg.jpg "); err != nil || tab.Background() != "https://img.example/bg.jpg" {
		t.Fatalf("%q %v", tab.Background(), err)
	}
}
