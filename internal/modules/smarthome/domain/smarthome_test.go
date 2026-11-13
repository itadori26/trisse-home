package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" http://homeassistant.local:8123/ ", " token ")
	if err != nil || s.URL != "http://homeassistant.local:8123" || s.Token != "token" {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	for _, c := range [][2]string{{"homeassistant.local", "t"}, {"http://ha", " "}} {
		if _, err := NewSettings(c[0], c[1]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%v: expected a validation error, got %v", c, err)
		}
	}
}

func TestEntityAction(t *testing.T) {
	cases := map[string]struct {
		kind ActionKind
		svc  Service
	}{
		"switch.coffee":       {ActionToggle, Service{"switch", "toggle"}},
		"light.kitchen":       {ActionToggle, Service{"light", "toggle"}},
		"button.doorbell":     {ActionPress, Service{"button", "press"}},
		"script.good_night":   {ActionPress, Service{"script", "turn_on"}},
		"sensor.power":        {ActionNone, Service{}},
		"lock.front_door":     {ActionNone, Service{}},
		"alarm_control_panel": {ActionNone, Service{}},
	}
	for id, want := range cases {
		kind, svc := Entity{ID: id}.Action()
		if kind != want.kind || svc != want.svc {
			t.Errorf("%s: got %q %+v", id, kind, svc)
		}
	}
}

func TestActionPolicy(t *testing.T) {
	p := ActionPolicy{Allowed: []string{"switch.coffee", "sensor.power"}}
	if svc, err := p.Check(Entity{ID: "switch.coffee"}); err != nil || svc.Name != "toggle" {
		t.Fatalf("allowed switch: %+v %v", svc, err)
	}
	if _, err := p.Check(Entity{ID: "switch.heater"}); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("not allow-listed: %v", err)
	}
	if _, err := p.Check(Entity{ID: "sensor.power"}); !errors.Is(err, ErrNoAction) {
		t.Fatalf("read-only entity: %v", err)
	}
}

func TestValidEntityIDAndSort(t *testing.T) {
	if !ValidEntityID("switch.coffee_1") || ValidEntityID("switch") || ValidEntityID("Switch.x") || ValidEntityID("a.b/c") {
		t.Fatal("entity ID validation")
	}
	list := []Entity{{ID: "b.x", Name: "beta"}, {ID: "a.x", Name: "Alpha"}}
	SortByName(list)
	if list[0].ID != "a.x" {
		t.Fatalf("sort: %+v", list)
	}
}
