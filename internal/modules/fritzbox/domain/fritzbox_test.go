package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

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

func TestCallFilter(t *testing.T) {
	calls := []Call{{Type: CallMissed, Number: "1"}, {Type: CallIncoming, Number: "2"}, {Type: CallOutgoing, Number: "3"},
		{Type: CallMissed, Number: "4"}, {Type: CallRejected, Number: "5"}}
	f, err := NewCallFilter("", 0)
	if err != nil || f.Kind() != CallsAll || f.Limit() != DefaultCalls || len(f.Apply(calls)) != 5 {
		t.Fatalf("defaults: %+v %v", f, err)
	}
	missed, _ := NewCallFilter("missed", 5)
	if got := missed.Apply(calls); len(got) != 2 || got[0].Number != "1" || got[1].Number != "4" {
		t.Fatalf("missed: %+v", got)
	}
	out, _ := NewCallFilter("outgoing", 5)
	if got := out.Apply(calls); len(got) != 1 || got[0].Number != "3" {
		t.Fatalf("outgoing: %+v", got)
	}
	five, _ := NewCallFilter("all", 5)
	many := append(append(calls, calls...), calls...)
	if got := five.Apply(many); len(got) != 5 {
		t.Fatalf("limit: %d", len(got))
	}
	for _, bad := range []struct {
		kind  string
		limit int
	}{{"rejected", 10}, {"all", 4}, {"all", 51}, {"MISSED", 10}} {
		if _, err := NewCallFilter(bad.kind, bad.limit); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%+v: want a validation error, got %v", bad, err)
		}
	}
}
