package domain

import "testing"

func TestVersion(t *testing.T) {
	cases := []struct {
		a, b  string
		newer bool
	}{
		{"1.2.4", "1.2.3", true},
		{"v1.10.0", "1.9.9", true},
		{"2.0.0", "1.99.99", true},
		{"1.2.3", "1.2.3", false},
		{"1.2.3", "v1.3.0", false},
		{"1.3.0", "dev", false},
		{"dev", "1.0.0", false},
		{"1.3.0-rc.1", "1.2.0", true},
	}
	for _, c := range cases {
		if got := ParseVersion(c.a).IsNewerThan(ParseVersion(c.b)); got != c.newer {
			t.Errorf("%s > %s: got %v", c.a, c.b, got)
		}
	}
	if ParseVersion("abc123").IsRelease() || !ParseVersion("v0.1.0").IsRelease() {
		t.Fatal("release detection wrong")
	}
}
