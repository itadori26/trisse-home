package domain

import "testing"

func TestNewSettings(t *testing.T) {
	s, err := NewSettings(" http://nas:8765/ ", " tok ", 100, 0)
	if err != nil || s.URL != "http://nas:8765" || s.Token != "tok" || !s.HasThresholds() {
		t.Fatalf("%+v %v", s, err)
	}
	for _, c := range []struct {
		url, token string
		down       float64
	}{{"ftp://x", "t", 0}, {"http://x", " ", 0}, {"http://x", "t", -1}} {
		if _, err := NewSettings(c.url, c.token, c.down, 0); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
}

func TestJudge(t *testing.T) {
	s := Settings{MinDownMbps: 100, MinUpMbps: 20}
	if ok, _ := s.Judge(Measurement{DownloadBps: 250e6, UploadBps: 40e6}); !ok {
		t.Fatal("fast is healthy")
	}
	ok, detail := s.Judge(Measurement{DownloadBps: 50e6, UploadBps: 10e6})
	if ok || detail != "Internet is slow: download 50 Mbit/s (minimum 100), upload 10 Mbit/s (minimum 20)" {
		t.Fatalf("%v %q", ok, detail)
	}
	if ok, _ := (Settings{}).Judge(Measurement{}); !ok {
		t.Fatal("no thresholds means healthy")
	}
}
