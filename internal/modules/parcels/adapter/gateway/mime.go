package gateway

import (
	"bytes"
	"encoding/base64"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/modules/parcels/domain"
)

var (
	hrefRe = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
	tagRe  = regexp.MustCompile(`(?s)<(?:style|script)[^>]*>.*?</(?:style|script)>|<[^>]+>`)
	words  = &mime.WordDecoder{CharsetReader: charsetReader}
)

// ParseMail turns the fetched header and text into what the extractors read:
// sender, subject and the plain text (HTML reduced to its text and link
// targets). Attachments and unknown parts are skipped; a body cut off at the
// fetch limit is read as far as it goes.
func ParseMail(header, text []byte) domain.Mail {
	msg, err := mail.ReadMessage(io.MultiReader(bytes.NewReader(header), strings.NewReader("\r\n"), bytes.NewReader(text)))
	if err != nil {
		return domain.Mail{Text: string(text)}
	}
	from, _ := words.DecodeHeader(msg.Header.Get("From"))
	subject, _ := words.DecodeHeader(msg.Header.Get("Subject"))
	var b strings.Builder
	collect(&b, textproto.MIMEHeader(msg.Header), msg.Body, 0)
	return domain.Mail{From: from, Subject: subject, Text: b.String()}
}

// collect appends the readable text of one MIME entity.
func collect(b *strings.Builder, h textproto.MIMEHeader, body io.Reader, depth int) {
	if depth > 4 || b.Len() > maxTextBytes {
		return
	}
	ct := h.Get("Content-Type")
	if ct == "" {
		ct = "text/plain"
	}
	mt, params, err := mime.ParseMediaType(ct)
	if err != nil {
		mt = "text/plain"
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(h.Get("Content-Disposition"))), "attachment") {
		return
	}
	switch {
	case strings.HasPrefix(mt, "multipart/"):
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				return // end, or a body cut off at the fetch limit
			}
			collect(b, part.Header, part, depth+1)
		}
	case mt == "text/plain" || mt == "text/html":
		raw, _ := io.ReadAll(io.LimitReader(decodeTransfer(h.Get("Content-Transfer-Encoding"), body), maxTextBytes))
		s := toUTF8(raw, params["charset"])
		if mt == "text/html" {
			s = htmlText(s)
		}
		b.WriteString(s)
		b.WriteString("\n")
	}
}

func decodeTransfer(enc string, r io.Reader) io.Reader {
	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "quoted-printable":
		return quotedprintable.NewReader(r)
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, &noSpace{r: r})
	}
	return r
}

// noSpace drops line breaks and spaces, which base64 bodies contain.
type noSpace struct{ r io.Reader }

func (n *noSpace) Read(p []byte) (int, error) {
	k, err := n.r.Read(p)
	j := 0
	for _, c := range p[:k] {
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			p[j] = c
			j++
		}
	}
	return j, err
}

// htmlText keeps the link targets (tracking links carry the number) and the
// visible text.
func htmlText(s string) string {
	var b strings.Builder
	for _, m := range hrefRe.FindAllStringSubmatch(s, -1) {
		b.WriteString(html.UnescapeString(m[1]))
		b.WriteString("\n")
	}
	b.WriteString(html.UnescapeString(tagRe.ReplaceAllString(s, "\n")))
	return b.String()
}

// toUTF8 converts Latin-1/Windows-1252 text; UTF-8 and ASCII stay.
func toUTF8(raw []byte, charset string) string {
	cs := strings.ToLower(charset)
	if utf8.Valid(raw) && !strings.HasPrefix(cs, "iso-8859") && !strings.HasPrefix(cs, "windows-125") {
		return string(raw)
	}
	r := make([]rune, len(raw))
	for i, c := range raw {
		r[i] = rune(c)
	}
	return string(r)
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	raw, err := io.ReadAll(io.LimitReader(input, 4096))
	if err != nil {
		return nil, err
	}
	return strings.NewReader(toUTF8(raw, charset)), nil
}
