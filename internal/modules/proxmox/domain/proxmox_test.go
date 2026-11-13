package domain

import (
	"errors"
	"strings"
	"testing"

	"trisse-home/internal/shared/kernel"
)

const fp = "AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89"

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" https://10.0.0.5:8006/ ", " monitor@pve!trisse-home ", " uuid ", fp)
	if err != nil {
		t.Fatal(err)
	}
	if s.URL != "https://10.0.0.5:8006" || s.TokenID != "monitor@pve!trisse-home" || s.Secret != "uuid" ||
		s.Fingerprint != strings.ToLower(strings.ReplaceAll(fp, ":", "")) {
		t.Fatalf("not normalised: %+v", s)
	}
	cases := map[string][4]string{
		"no scheme":            {"10.0.0.5:8006", "a@pve!t", "s", ""},
		"path":                 {"https://pve/api2/json", "a@pve!t", "s", ""},
		"token without realm":  {"https://pve", "monitor!t", "s", ""},
		"token without name":   {"https://pve", "monitor@pve", "s", ""},
		"no secret":            {"https://pve", "a@pve!t", " ", ""},
		"short fingerprint":    {"https://pve", "a@pve!t", "s", "AB:CD"},
		"fingerprint and http": {"http://pve", "a@pve!t", "s", fp},
	}
	for name, c := range cases {
		if _, err := NewSettings(c[0], c[1], c[2], c[3]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: expected a validation error, got %v", name, err)
		}
	}
}

func TestClusterSortAndLookup(t *testing.T) {
	c := Cluster{
		Nodes:  []Node{{Name: "pve2"}, {Name: "pve1"}},
		Guests: []Guest{{ID: 105, Name: "b"}, {ID: 100, Name: "a", Status: "running"}},
	}
	c.Sort()
	if c.Nodes[0].Name != "pve1" || c.Guests[0].ID != 100 {
		t.Fatalf("sort: %+v", c)
	}
	if g, ok := c.Guest(100); !ok || !g.Running() {
		t.Fatalf("lookup: %+v %v", g, ok)
	}
	if _, ok := c.Guest(999); ok {
		t.Fatal("unknown guest found")
	}
}

func TestActionPolicy(t *testing.T) {
	p := ActionPolicy{Allowed: []int{100}}
	if err := p.Check(100, ActionStart); err != nil {
		t.Fatal(err)
	}
	if err := p.Check(100, "stop"); !errors.Is(err, ErrUnknownAction) {
		t.Fatalf("hard stop must not be possible: %v", err)
	}
	if err := p.Check(101, ActionShutdown); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("not allow-listed: %v", err)
	}
	if (ActionPolicy{}).Allows(100) {
		t.Fatal("actions are off without an allow-list")
	}
}
