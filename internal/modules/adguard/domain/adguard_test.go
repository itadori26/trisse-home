package domain

import "testing"

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" http://10.0.0.2:3000/ ", "admin", "secret")
	if err != nil || s.URL != "http://10.0.0.2:3000" {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	for _, tc := range [][3]string{{"10.0.0.2", "a", "b"}, {"http://x", "", "b"}, {"http://x", "a", ""}} {
		if _, err := NewSettings(tc[0], tc[1], tc[2]); err == nil {
			t.Errorf("expected error for %v", tc)
		}
	}
}

func TestBlockedPercent(t *testing.T) {
	if (Stats{}).BlockedPercent() != 0 || (Stats{Queries: 200, Blocked: 50}).BlockedPercent() != 25 {
		t.Fatal("percent wrong")
	}
}
