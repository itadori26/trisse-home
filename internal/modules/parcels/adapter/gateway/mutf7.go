package gateway

import (
	"encoding/base64"
	"strings"
	"unicode/utf16"
)

// mailboxEncoding is base64 with "," instead of "/" and without padding
// (RFC 3501, 5.1.3).
var mailboxEncoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+,").WithPadding(base64.NoPadding)

// encodeMailbox writes a folder name in IMAP's modified UTF-7, so names
// such as "Päckchen" reach the server: printable ASCII stays, "&" becomes
// "&-", other characters become "&" + base64(UTF-16BE) + "-".
func encodeMailbox(name string) string {
	var b strings.Builder
	var run []rune
	flush := func() {
		if len(run) == 0 {
			return
		}
		units := utf16.Encode(run)
		raw := make([]byte, 0, len(units)*2)
		for _, u := range units {
			raw = append(raw, byte(u>>8), byte(u))
		}
		b.WriteByte('&')
		b.WriteString(mailboxEncoding.EncodeToString(raw))
		b.WriteByte('-')
		run = run[:0]
	}
	for _, r := range name {
		switch {
		case r == '&':
			flush()
			b.WriteString("&-")
		case r >= 0x20 && r <= 0x7e:
			flush()
			b.WriteRune(r)
		default:
			run = append(run, r)
		}
	}
	flush()
	return b.String()
}
