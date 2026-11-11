package controller

import (
	"html"
	"io"
	"regexp"
	"strings"

	"trisse-home/internal/modules/bookmarks/domain"
)

// token matches the parts of the Netscape bookmark file format that matter:
// folder headings, list open/close, links and descriptions.
var (
	token     = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>|<dl[^>]*>|</dl>|<a\s([^>]*)>(.*?)</a>|<dd>([^<]*)`)
	attribute = regexp.MustCompile(`(?is)\b(href|tags)\s*=\s*"([^"]*)"`)
	tagsSplit = regexp.MustCompile(`[,;]`)
)

// ParseNetscape reads a browser bookmark export (Netscape HTML format, used
// by Firefox, Chrome, Edge and Safari). Nested folders become paths such as
// "Work / Docs".
func ParseNetscape(r io.Reader) ([]domain.ImportCandidate, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var out []domain.ImportCandidate
	var stack []string
	pending, havePending := "", false
	for _, m := range token.FindAllStringSubmatch(string(raw), -1) {
		lower := strings.ToLower(m[0])
		switch {
		case strings.HasPrefix(lower, "<h3"):
			pending, havePending = strings.TrimSpace(html.UnescapeString(stripTags(m[1]))), true
		case strings.HasPrefix(lower, "<dl"):
			if havePending {
				stack = append(stack, pending)
			} else {
				stack = append(stack, "") // the root list
			}
			havePending = false
		case strings.HasPrefix(lower, "</dl"):
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case strings.HasPrefix(lower, "<a"):
			c := domain.ImportCandidate{Title: strings.TrimSpace(html.UnescapeString(stripTags(m[3]))), FolderPath: path(stack)}
			for _, a := range attribute.FindAllStringSubmatch(m[2], -1) {
				v := html.UnescapeString(a[2])
				switch strings.ToLower(a[1]) {
				case "href":
					c.URL = strings.TrimSpace(v)
				case "tags":
					for _, t := range tagsSplit.Split(v, -1) {
						if t = strings.TrimSpace(t); t != "" {
							c.Tags = append(c.Tags, strings.ReplaceAll(t, " ", "-"))
						}
					}
				}
			}
			out = append(out, c)
		case strings.HasPrefix(lower, "<dd") && len(out) > 0:
			out[len(out)-1].Description = strings.TrimSpace(html.UnescapeString(m[4]))
		}
	}
	return out, nil
}

var anyTag = regexp.MustCompile(`<[^>]*>`)

func stripTags(s string) string { return anyTag.ReplaceAllString(s, "") }

func path(stack []string) string {
	var parts []string
	for _, p := range stack {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " / ")
}
