package kernel

import (
	"errors"
	"net/netip"
	"testing"
)

func TestParseTarget(t *testing.T) {
	ok := map[string]string{
		"https://nas.local:8443":     "https://nas.local:8443",
		"http://192.168.178.10":      "http://192.168.178.10",
		"http://127.0.0.1:8080":      "http://127.0.0.1:8080", // loopback stays allowed
		"https://[fd00::1]:8006":     "https://[fd00::1]:8006",
		"https://10.0.0.2":           "https://10.0.0.2",
		"https://pve.example.org/":   "https://pve.example.org",
		"http://[::ffff:10.0.0.1]:1": "http://[::ffff:10.0.0.1]:1",
	}
	ok["\thttps://nas.local:8443/\n"] = "https://nas.local:8443" // surrounding white space is trimmed
	for raw, want := range ok {
		u, err := ParseTarget("url", raw)
		if err != nil || u.String() != want {
			t.Errorf("%q: got %v %v, want %s", raw, u, err, want)
		}
	}
	bad := []string{
		"", "nas.local", "ftp://nas.local", "file:///etc/passwd", "https://", "javascript:alert(1)",
		"https://user:pass@nas.local", // credentials
		"https://nas.local?x=1",       // query
		"https://nas.local/#frag",     // fragment
		"https://nas.local/?",         // empty query
		"https://nas.local/admin",     // path without permission
		"http://169.254.169.254",      // cloud metadata
		"http://[fe80::1]:80",         // IPv6 link-local
		"http://[::ffff:169.254.1.1]", // mapped link-local
		"http://0.0.0.0:80",           // unspecified
		"http://224.0.0.1",            // multicast
		"http://0.1.2.3",              // "this network"
	}
	for _, raw := range bad {
		if _, err := ParseTarget("url", raw); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q accepted: %v", raw, err)
		}
	}
}

func TestParseTargetWithPath(t *testing.T) {
	u, err := ParseTarget("url", "https://cloud.example/remote.php/dav/files/me/backups/", WithTargetPath())
	if err != nil || u.String() != "https://cloud.example/remote.php/dav/files/me/backups" {
		t.Fatalf("%v %v", u, err)
	}
	if _, err := ParseTarget("url", "https://cloud.example/dav?x", WithTargetPath()); err == nil {
		t.Fatal("query accepted with a path")
	}
	var v *ValidationError
	if _, err := ParseTarget("serverUrl", "https://x/y"); !errors.As(err, &v) || v.Field != "serverUrl" {
		t.Fatalf("field: %v", err)
	}
}

func TestTargetAddrAllowed(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1": true, "::1": true, "192.168.1.1": true, "10.13.13.2": true, "fd00::1": true, "93.184.216.34": true,
		"169.254.169.254": false, "fe80::1": false, "::": false, "0.0.0.0": false, "ff02::1": false, "::ffff:169.254.0.1": false,
	} {
		if got := TargetAddrAllowed(netip.MustParseAddr(addr)); got != want {
			t.Errorf("%s: got %v", addr, got)
		}
	}
}

func TestTargetChanged(t *testing.T) {
	if TargetChanged([]string{"https://pve:8006/", "root@pam!th"}, []string{"https://PVE:8006", "root@pam!th"}) {
		t.Fatal("trailing slash and case are no change")
	}
	if !TargetChanged([]string{"https://pve:8006", "a"}, []string{"https://other:8006", "a"}) {
		t.Fatal("new host is a change")
	}
	if !TargetChanged([]string{"x"}, []string{"x", "y"}) {
		t.Fatal("different length is a change")
	}
}
