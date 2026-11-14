package gateway

import "testing"

// Examples from RFC 3501 and common German folder names.
func TestEncodeMailbox(t *testing.T) {
	for in, want := range map[string]string{
		"INBOX":              "INBOX",
		"Pakete":             "Pakete",
		"Päckchen":           "P&AOQ-ckchen",
		"Entwürfe & Ideen":   "Entw&APw-rfe &- Ideen",
		"~peter/mail/台北/日本語": "~peter/mail/&U,BTFw-/&ZeVnLIqe-",
	} {
		if got := encodeMailbox(in); got != want {
			t.Errorf("encodeMailbox(%q) = %q, want %q", in, got, want)
		}
	}
}
