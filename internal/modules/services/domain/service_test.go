package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	s, err := NewService("1", " Nextcloud ", "https://cloud.example.com", "", "☁️", "", false, 0)
	if err != nil || s.Name != "Nextcloud" || s.CheckURL != s.URL {
		t.Fatalf("unexpected %+v %v", s, err)
	}
	for _, tc := range []struct{ name, url, check string }{
		{"", "https://x.y", ""},
		{"x", "ftp://x.y", ""},
		{"x", "https://x.y", "javascript:alert(1)"},
	} {
		if _, err := NewService("1", tc.name, tc.url, tc.check, "", "", false, 0); err == nil {
			t.Errorf("expected error for %+v", tc)
		}
	}
}

func TestStatusFromResponse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		code int
		err  error
		want State
	}{
		{200, nil, StateOnline},
		{401, nil, StateOnline},
		{502, nil, StateOffline},
		{0, errors.New("refused"), StateOffline},
	}
	for _, tt := range tests {
		if got := StatusFromResponse("1", tt.code, time.Millisecond, tt.err, now); got.State != tt.want {
			t.Errorf("code %d err %v: got %s want %s", tt.code, tt.err, got.State, tt.want)
		}
	}
}
