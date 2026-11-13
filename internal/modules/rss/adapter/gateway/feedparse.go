package gateway

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// A small feed parser for RSS 0.9x/2.0, RSS 1.0 (RDF), Atom 1.0 and JSON Feed.
// It reads only what the reader needs (feed title and site link, and per entry
// the id, title, link and date) and replaces the much larger gofeed library
// (about 1 MB of the binary, much of it character set tables for Chinese,
// Japanese and Korean). Of the character sets only UTF-8 and the Western and
// Central European ones are known. Differences to gofeed (checked against its
// own 460 test feeds): titles with markup come back as plain text, and of
// several links in the channel the first one is used.

var errNotAFeed = errors.New("not an RSS, Atom or JSON feed")

type parsedFeed struct {
	Title, Link string
	Items       []parsedItem
}

type parsedItem struct {
	ID, Title, Link string
	Published       time.Time
}

// parseFeed detects the format and parses it.
func parseFeed(r io.Reader) (parsedFeed, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return parsedFeed{}, err
	}
	trimmed := bytes.TrimLeft(bytes.TrimPrefix(body, []byte("\xef\xbb\xbf")), " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '{' {
		return parseJSONFeed(trimmed)
	}
	return parseXMLFeed(body)
}

// ---- XML -------------------------------------------------------------------

// node is a light element tree: real feeds are small, and a tree lets the
// extraction ignore elements of foreign namespaces (itunes:title, media:title)
// and read titles that contain markup.
type node struct {
	space, name string
	attrs       []xml.Attr
	kids        []*node // elements and, with name "#text", the text between them
	text        string  // content of a "#text" node
	base        string  // xml:base in effect for this element ("" if none)
}

const (
	nsAtom   = "http://www.w3.org/2005/Atom"
	nsAtom03 = "http://purl.org/atom/ns#"
	nsRSS1   = "http://purl.org/rss/1.0/"
	nsRSS09  = "http://my.netscape.com/rdf/simple/0.9/"
	nsRSS09b = "http://channel.netscape.com/rdf/simple/0.9/"
)

// dropIllegal removes the control characters XML forbids; feeds from careless
// generators contain them and would otherwise be rejected as a whole.
func dropIllegal(b []byte) []byte {
	legal := func(c byte) bool { return c >= 0x20 || c == '\t' || c == '\n' || c == '\r' }
	if !slices.ContainsFunc(b, func(c byte) bool { return !legal(c) }) {
		return b
	}
	out := make([]byte, 0, len(b))
	for _, c := range b {
		if legal(c) {
			out = append(out, c)
		}
	}
	return out
}

func parseXMLFeed(body []byte) (parsedFeed, error) {
	d := xml.NewDecoder(bytes.NewReader(dropIllegal(body)))
	d.Strict = false // tolerate what real feeds contain: stray ampersands, unclosed tags
	// Not xml.HTMLAutoClose: it lists <link>, which would be closed before its text.
	d.AutoClose = []string{"br", "hr", "img", "input"}
	d.Entity = xml.HTMLEntity
	d.CharsetReader = charsetReader
	var root *node
	var stack []*node
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if root == nil {
				return parsedFeed{}, errNotAFeed
			}
			break // a truncated feed still shows what arrived
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &node{space: t.Name.Space, name: strings.ToLower(t.Name.Local), attrs: t.Attr}
			if len(stack) > 0 {
				top := stack[len(stack)-1]
				n.base = top.base
				top.kids = append(top.kids, n)
			} else if root == nil {
				root = n
			}
			for _, a := range t.Attr {
				if a.Name.Local == "base" && (a.Name.Space == "xml" || a.Name.Space == "http://www.w3.org/XML/1998/namespace") {
					n.base = resolveURL(n.base, strings.TrimSpace(a.Value))
				}
			}
			if len(stack) < 64 {
				stack = append(stack, n)
			} else {
				return parsedFeed{}, errNotAFeed
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) > 0 {
				top := stack[len(stack)-1]
				top.kids = append(top.kids, &node{name: "#text", text: string(t)})
			}
		}
	}
	if root == nil {
		return parsedFeed{}, errNotAFeed
	}
	switch root.name {
	case "rss", "rdf":
		return fromRSS(root), nil
	case "feed":
		return fromAtom(root), nil
	}
	return parsedFeed{}, errNotAFeed // html, other xml
}

// core reports whether n belongs to the feed format itself and not to an
// extension: no namespace, the format's own namespaces, or the bogus default
// namespace some feeds declare on the root.
func (n *node) core(rootSpace string) bool {
	switch n.space {
	case "", rootSpace, nsRSS1, nsRSS09, nsRSS09b, nsAtom, nsAtom03:
		return true
	}
	return false
}

// child returns the first non-empty child that is core, or an empty one.
func (n *node) child(rootSpace string, names ...string) *node {
	for _, name := range names {
		for _, k := range n.kids {
			if k.name == name && k.core(rootSpace) && k.value() != "" {
				return k
			}
		}
	}
	return &node{}
}

// elem returns the first core child with the name (structure, no text needed).
func (n *node) elem(rootSpace, name string) *node {
	for _, k := range n.kids {
		if k.name == name && k.core(rootSpace) {
			return k
		}
	}
	return &node{}
}

func (n *node) children(rootSpace, name string) []*node {
	var out []*node
	for _, k := range n.kids {
		if k.name == name && k.core(rootSpace) {
			out = append(out, k)
		}
	}
	return out
}

// value is the text of the element and of everything inside it (a title may
// contain inline markup).
func (n *node) value() string {
	var b strings.Builder
	var walk func(*node)
	walk = func(n *node) {
		for _, k := range n.kids {
			if k.name == "#text" {
				b.WriteString(k.text)
			} else {
				walk(k)
			}
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}

func (n *node) attr(name string) string {
	for _, a := range n.attrs {
		if strings.EqualFold(a.Name.Local, name) && (a.Name.Space == "" || a.Name.Space == n.space) {
			return strings.TrimSpace(a.Value)
		}
	}
	return ""
}

// dcDate reads dc:date, the date RSS 1.0 feeds use.
func (n *node) dcDate() string {
	for _, k := range n.kids {
		if k.name == "date" && (k.space == "dc" || strings.Contains(k.space, "purl.org/dc/")) {
			return k.value()
		}
	}
	return ""
}

func fromRSS(root *node) parsedFeed {
	rs := root.space
	channel := root.elem(rs, "channel")
	out := parsedFeed{Title: channel.child(rs, "title").value(), Link: channel.child(rs, "link").value()}
	items := append(channel.children(rs, "item"), root.children(rs, "item")...) // RSS 1.0: items beside the channel
	for _, it := range items {
		link := resolveURL(firstNonEmpty(it.base, channel.base, out.Link), it.child(rs, "link").value())
		id := it.child(rs, "guid").value()
		if id == "" {
			id = link
		}
		out.Items = append(out.Items, parsedItem{
			ID: id, Title: it.child(rs, "title").value(), Link: link,
			Published: firstDate(it.child(rs, "pubdate").value(), it.dcDate(),
				it.child(rs, "published").value(), it.child(rs, "updated").value()),
		})
	}
	return out
}

func fromAtom(root *node) parsedFeed {
	rs := root.space
	out := parsedFeed{Title: root.child(rs, "title").value(), Link: alternate(root.children(rs, "link"), root.base)}
	for _, e := range root.children(rs, "entry") {
		link := alternate(e.children(rs, "link"), firstNonEmpty(e.base, root.base, out.Link))
		id := e.child(rs, "id").value()
		if id == "" {
			id = link
		}
		out.Items = append(out.Items, parsedItem{
			ID: id, Title: e.child(rs, "title").value(), Link: link,
			Published: firstDate(e.child(rs, "published").value(), e.child(rs, "updated").value(),
				e.child(rs, "issued").value(), e.child(rs, "modified").value(), e.child(rs, "created").value()),
		})
	}
	return out
}

// alternate picks the link to the page: rel="alternate" or no rel, else the
// first link that is not the feed itself.
func alternate(links []*node, fallbackBase string) string {
	var other string
	for _, l := range links {
		href := l.attr("href")
		if href == "" {
			continue
		}
		href = resolveURL(firstNonEmpty(l.base, fallbackBase), href)
		switch strings.ToLower(l.attr("rel")) {
		case "", "alternate":
			return href
		case "self", "enclosure", "edit", "replies":
		default:
			if other == "" {
				other = href
			}
		}
	}
	return other
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// resolveURL makes ref absolute against base (an xml:base or the site link).
// Without a usable base, or if ref already is absolute, ref is returned as it is.
func resolveURL(base, ref string) string {
	r, err := url.Parse(ref)
	if err != nil || r.IsAbs() || base == "" {
		return ref
	}
	b, err := url.Parse(base)
	if err != nil || !b.IsAbs() {
		return ref
	}
	return b.ResolveReference(r).String()
}

// charsetReader turns the declared encoding into UTF-8. Unknown names fail,
// so such a feed is reported as invalid instead of showing garbage.
func charsetReader(label string, input io.Reader) (io.Reader, error) {
	var enc encoding.Encoding
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "iso-8859-1", "iso8859-1", "latin1", "latin-1", "l1", "us-ascii", "ascii", "iso_8859-1":
		enc = charmap.ISO8859_1
	case "windows-1252", "cp1252", "x-cp1252":
		enc = charmap.Windows1252
	case "iso-8859-15", "latin9", "latin-9":
		enc = charmap.ISO8859_15
	case "iso-8859-2", "latin2":
		enc = charmap.ISO8859_2
	case "windows-1250", "cp1250":
		enc = charmap.Windows1250
	case "windows-1251", "cp1251":
		enc = charmap.Windows1251
	default:
		return nil, fmt.Errorf("unsupported character set %q", label)
	}
	return enc.NewDecoder().Reader(input), nil
}

// ---- JSON Feed -------------------------------------------------------------

type jsonFeed struct {
	Title string `json:"title"`
	Home  string `json:"home_page_url"`
	Items []struct {
		ID        string `json:"id"`
		URL       string `json:"url"`
		Title     string `json:"title"`
		Published string `json:"date_published"`
		Modified  string `json:"date_modified"`
	} `json:"items"`
}

func parseJSONFeed(body []byte) (parsedFeed, error) {
	var f jsonFeed
	if err := json.Unmarshal(body, &f); err != nil || (f.Title == "" && f.Items == nil) {
		return parsedFeed{}, errNotAFeed
	}
	out := parsedFeed{Title: strings.TrimSpace(f.Title), Link: strings.TrimSpace(f.Home)}
	for _, it := range f.Items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			id = strings.TrimSpace(it.URL)
		}
		out.Items = append(out.Items, parsedItem{
			ID: id, Title: strings.TrimSpace(it.Title), Link: strings.TrimSpace(it.URL),
			Published: firstDate(it.Published, it.Modified),
		})
	}
	return out, nil
}

// ---- dates -----------------------------------------------------------------

var dateLayouts = []string{
	time.RFC3339Nano, time.RFC3339,
	time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822,
	"Mon, 2 Jan 2006 15:04:05 -0700", "Mon, 2 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 2006 15:04 -0700", "Mon, 2 Jan 2006 15:04 MST",
	"Mon, 02 Jan 2006 15:04:05", "Mon, 2 Jan 2006 15:04:05",
	"2 Jan 2006 15:04:05 -0700", "2 Jan 2006 15:04:05 MST", "2 Jan 2006 15:04 -0700",
	"Mon Jan 2 15:04:05 2006", "Mon Jan 2 15:04:05 MST 2006", "Mon Jan 2 15:04:05 -0700 2006",
	"2006-01-02T15:04:05", "2006-01-02 15:04:05 -0700", "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05",
	"2006-01-02T15:04Z07:00", "2006-01-02",
}

// zones maps the abbreviations feeds use to offsets; Go would silently treat
// them as UTC.
var zones = map[string]string{
	"UT": "+0000", "GMT": "+0000", "Z": "+0000", "UTC": "+0000",
	"EST": "-0500", "EDT": "-0400", "CST": "-0600", "CDT": "-0500",
	"MST": "-0700", "MDT": "-0600", "PST": "-0800", "PDT": "-0700",
	"CET": "+0100", "CEST": "+0200",
}

func firstDate(values ...string) time.Time {
	for _, v := range values {
		if t, ok := parseDate(v); ok {
			return t
		}
	}
	return time.Time{}
}

func parseDate(s string) (time.Time, bool) {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return time.Time{}, false
	}
	if i := strings.LastIndexByte(s, ' '); i > 0 {
		if off, ok := zones[strings.ToUpper(s[i+1:])]; ok {
			s = s[:i+1] + off
		}
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
