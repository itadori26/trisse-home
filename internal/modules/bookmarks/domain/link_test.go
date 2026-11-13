package domain

import (
	"strings"
	"testing"
	"time"
)

func TestClassifyLink(t *testing.T) {
	now := time.Unix(100, 0)
	cases := []struct {
		name         string
		code         int
		final        string
		failed       bool
		want         LinkState
		wantTarget   string
		wantZeroCode bool
	}{
		{"ok", 200, "https://go.dev/", false, LinkOK, "", false},
		{"same URL with slash and port", 200, "https://go.dev:443", false, LinkOK, "", false},
		{"redirect", 200, "https://go.dev/doc/", false, LinkRedirect, "https://go.dev/doc/", false},
		{"login required", 401, "", false, LinkOK, "", false},
		{"forbidden", 403, "", false, LinkOK, "", false},
		{"rate limited", 429, "", false, LinkOK, "", false},
		{"not found", 404, "", false, LinkBroken, "", false},
		{"server error", 502, "", false, LinkBroken, "", false},
		{"network error", 0, "", true, LinkBroken, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := ClassifyLink(tc.code, "https://go.dev", tc.final, tc.failed, now)
			if s.State != tc.want || s.Target != tc.wantTarget || !s.CheckedAt.Equal(now) || (tc.wantZeroCode && s.Code != 0) {
				t.Fatalf("got %+v", s)
			}
		})
	}
}

func TestLinkStatusDue(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	if !(LinkStatus{}).Due(now) {
		t.Fatal("unchecked links are due")
	}
	if (LinkStatus{CheckedAt: now.Add(-LinkCheckInterval + time.Minute)}).Due(now) {
		t.Fatal("checked within a week")
	}
	if !(LinkStatus{CheckedAt: now.Add(-LinkCheckInterval)}).Due(now) {
		t.Fatal("a week old check is due")
	}
}

func TestIconAndTitleRules(t *testing.T) {
	if _, err := NewIcon("image/png; charset=binary", []byte{1}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []struct {
		ct   string
		size int
	}{{"image/svg+xml", 10}, {"text/html", 10}, {"image/png", 0}, {"image/png", MaxIconBytes + 1}} {
		if _, err := NewIcon(bad.ct, make([]byte, bad.size)); err == nil {
			t.Errorf("accepted %s (%d bytes)", bad.ct, bad.size)
		}
	}
	if got := CleanTitle("  Go\n\t Docs  "); got != "Go Docs" {
		t.Fatalf("got %q", got)
	}
	if got := CleanTitle(strings.Repeat("x", 300)); len([]rune(got)) != MaxTitleLength {
		t.Fatalf("not shortened: %d", len([]rune(got)))
	}
	for declared, want := range map[string]string{
		"":                          "https://go.dev/favicon.ico",
		"/images/icon.png":          "https://go.dev/images/icon.png",
		"icon.png":                  "https://go.dev/doc/icon.png",
		"https://cdn.example/i.png": "https://cdn.example/i.png",
		"data:image/png;base64,AA":  "https://go.dev/favicon.ico",
	} {
		if got := ResolveIconURL("https://go.dev/doc/page", declared); got != want {
			t.Errorf("%q: got %q, want %q", declared, got, want)
		}
	}
}
