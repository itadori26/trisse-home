package domain

import "testing"

func TestNewRemoteSettings(t *testing.T) {
	s, err := NewRemoteSettings(" https://cloud.example.org/remote.php/dav/files/me/Backups ", " me ", "app-pw", 10)
	if err != nil || s.URL != "https://cloud.example.org/remote.php/dav/files/me/Backups/" || s.Username != "me" || s.RemotePolicy().Retention != 10 {
		t.Fatalf("%+v %v", s, err)
	}
	for _, c := range []struct {
		url, user, pass string
		keep            int
	}{
		{"ftp://x/", "u", "p", 1},
		{"https://x/?a=1", "u", "p", 1},
		{"https://x/", "", "p", 1},
		{"https://x/", "u", "", 1},
		{"https://x/", "u", "p", 0},
		{"https://x/", "u", "p", 366},
	} {
		if _, err := NewRemoteSettings(c.url, c.user, c.pass, c.keep); err == nil {
			t.Errorf("accepted %+v", c)
		}
	}
}
