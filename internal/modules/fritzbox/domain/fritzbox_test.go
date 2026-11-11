package domain

import "testing"

func TestSettings(t *testing.T) {
	s, err := NewSettings("http://fritz.box", "", "")
	if err != nil || s.URL != "http://fritz.box:49000" || s.HasLogin() {
		t.Fatalf("%+v %v", s, err)
	}
	if s, _ := NewSettings("", "home", "pw"); s.URL != DefaultURL || !s.HasLogin() {
		t.Fatalf("default URL: %+v", s)
	}
	for _, bad := range []string{"https://fritz.box", "ftp://x", "http://fritz.box/path"} {
		if _, err := NewSettings(bad, "", ""); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
	if _, err := NewSettings("", "home", ""); err == nil {
		t.Fatal("user without password accepted")
	}
}

func TestWANPercentAndCallTypes(t *testing.T) {
	w := WAN{DownstreamBs: 12_500_000, DownstreamMaxBits: 200_000_000}
	if p := w.DownstreamPercent(); p != 50 {
		t.Fatalf("percent %v", p)
	}
	if (WAN{UpstreamBs: 5}).UpstreamPercent() != 0 {
		t.Fatal("unknown link speed must give 0")
	}
	if CallTypeFromAVM(2) != CallMissed || CallTypeFromAVM(3) != CallOutgoing || CallTypeFromAVM(11) != CallActive {
		t.Fatal("call type mapping wrong")
	}
}
