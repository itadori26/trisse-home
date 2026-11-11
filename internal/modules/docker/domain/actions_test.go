package domain

import "testing"

func TestActionPolicy(t *testing.T) {
	p := ActionPolicy{Enabled: true, Allowed: []string{"homarr", "dashdot"}}
	if !p.Allows("homarr") || p.Allows("nextcloud-app") || p.Allows("../homarr") || p.Allows("") {
		t.Fatal("allow-list not enforced")
	}
	p.Enabled = false
	if p.Allows("homarr") || len(p.AllowedNames()) != 0 || p.Check("homarr") == nil {
		t.Fatal("disabled policy must allow nothing")
	}
}
