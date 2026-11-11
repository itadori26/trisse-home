package domain

import (
	"testing"
	"time"
)

func TestNewAPIToken(t *testing.T) {
	now := time.Unix(1000, 0)
	tok, err := NewAPIToken("1", " Shortcuts ", "h", []string{"todos:write", "TODOS:write", "notes:read"}, now, time.Time{})
	if err != nil || tok.Name != "Shortcuts" || len(tok.Scopes) != 2 || tok.Expired(now.AddDate(10, 0, 0)) {
		t.Fatalf("%+v %v", tok, err)
	}
	for _, scopes := range [][]string{nil, {"todos"}, {"auth:write"}, {"todos:admin"}, {"../x:read"}} {
		if _, err := NewAPIToken("1", "x", "h", scopes, now, time.Time{}); err == nil {
			t.Errorf("scopes %v accepted", scopes)
		}
	}
	if _, err := NewAPIToken("1", "x", "h", []string{"todos:read"}, now, now); err == nil {
		t.Fatal("expiry in the past accepted")
	}
	exp, _ := NewAPIToken("1", "x", "h", []string{"todos:read"}, now, now.Add(time.Hour))
	if exp.Expired(now) || !exp.Expired(now.Add(time.Hour)) {
		t.Fatal("expiry wrong")
	}
}

func TestTokenMayAccess(t *testing.T) {
	scopes := []string{"todos:write", "notes:read"}
	cases := []struct {
		method, path string
		want         bool
	}{
		{"POST", "/api/todos", true},
		{"GET", "/api/todos/1", true},
		{"GET", "/api/notes", true},
		{"PUT", "/api/notes/1", false},
		{"GET", "/api/bookmarks", false},
		{"GET", "/api/auth/tokens", false},
		{"POST", "/api/auth/password", false},
		{"GET", "/healthz", false},
	}
	for _, c := range cases {
		if got := TokenMayAccess(scopes, c.method, c.path); got != c.want {
			t.Errorf("%s %s: %v, want %v", c.method, c.path, got, c.want)
		}
	}
}

func TestRecordUseIsThrottled(t *testing.T) {
	var tok APIToken
	now := time.Unix(0, 0)
	if !tok.RecordUse(now, "1.1.1.1") || tok.RecordUse(now.Add(time.Minute), "1.1.1.1") {
		t.Fatal("second use within the hour must not be persisted")
	}
	if !tok.RecordUse(now.Add(2*time.Minute), "2.2.2.2") || !tok.RecordUse(now.Add(3*time.Hour), "2.2.2.2") {
		t.Fatal("new IP or an hour later must be persisted")
	}
}
