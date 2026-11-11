package ical

import "strings"

const maxLineOctets = 75

// EscapeText escapes a TEXT value (RFC 5545 §3.3.11).
func EscapeText(s string) string {
	return strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`, "\r", "").Replace(s)
}

// FoldLine splits a content line into chunks of at most 75 octets without
// breaking UTF-8 sequences and terminates it with CRLF (RFC 5545 §3.1).
func FoldLine(line string) string {
	var b strings.Builder
	limit := maxLineOctets
	for len(line) > limit {
		cut := limit
		for cut > 0 && line[cut]&0xC0 == 0x80 { // continuation byte: move back
			cut--
		}
		b.WriteString(line[:cut] + "\r\n ")
		line = line[cut:]
		limit = maxLineOctets - 1 // the leading space counts
	}
	b.WriteString(line + "\r\n")
	return b.String()
}
