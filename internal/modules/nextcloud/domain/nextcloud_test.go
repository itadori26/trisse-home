package domain

import "testing"

func TestSettingsAndQuota(t *testing.T) {
	s, err := NewSettings(" https://cloud.example.com/ ", "anna", "app-pw")
	if err != nil || s.URL != "https://cloud.example.com" {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := NewSettings("cloud.example.com", "anna", "x"); err == nil {
		t.Fatal("URL without scheme accepted")
	}
	if _, err := NewSettings("https://c", "anna", ""); err == nil {
		t.Fatal("missing password accepted")
	}
	if p := (Quota{Used: 25, Total: 100}).Percent(); p != 25 {
		t.Fatalf("percent %v", p)
	}
	if p := (Quota{Used: 25}).Percent(); p != 0 {
		t.Fatal("unlimited quota must be 0 %")
	}
	if ServerInfoHint(404) == ServerInfoHint(403) {
		t.Fatal("hints must differ")
	}
}
