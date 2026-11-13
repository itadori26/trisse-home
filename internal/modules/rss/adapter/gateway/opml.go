package gateway

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/rss/domain"
)

// OPML implements usecase.OPMLCodec (OPML 2.0, #97).
//
// encoding/xml never resolves external entities or DTDs, so XXE and
// entity expansion are not possible; the size is limited by the caller
// (domain.MaxImportBytes) and the nesting depth here.
type OPML struct{}

type opmlDoc struct {
	XMLName xml.Name      `xml:"opml"`
	Version string        `xml:"version,attr"`
	Head    opmlHead      `xml:"head"`
	Body    []opmlOutline `xml:"body>outline"`
}

type opmlHead struct {
	Title       string `xml:"title"`
	DateCreated string `xml:"dateCreated,omitempty"`
}

type opmlOutline struct {
	Text     string        `xml:"text,attr"`
	Title    string        `xml:"title,attr,omitempty"`
	Type     string        `xml:"type,attr,omitempty"`
	XMLURL   string        `xml:"xmlUrl,attr,omitempty"`
	HTMLURL  string        `xml:"htmlUrl,attr,omitempty"`
	Children []opmlOutline `xml:"outline"`
}

const maxDepth = 8

var errNotOPML = errors.New("not an OPML file")

// Decode implements usecase.OPMLCodec: every outline with an xmlUrl is a
// feed; outlines without one are folders whose names become the category
// of the feeds inside ("Tech / Go" for nested folders).
func (OPML) Decode(data []byte) ([]domain.ImportEntry, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = true
	dec.Entity = nil // only the five predefined XML entities
	var doc opmlDoc
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("%w: %w", errNotOPML, err)
	}
	var out []domain.ImportEntry
	var walk func(list []opmlOutline, path []string, depth int)
	walk = func(list []opmlOutline, path []string, depth int) {
		if depth > maxDepth {
			return
		}
		for _, o := range list {
			name := strings.TrimSpace(o.Title)
			if name == "" {
				name = strings.TrimSpace(o.Text)
			}
			if strings.TrimSpace(o.XMLURL) != "" {
				out = append(out, domain.ImportEntry{Title: name, URL: o.XMLURL, SiteURL: o.HTMLURL, Category: strings.Join(path, " / ")})
				continue
			}
			walk(o.Children, append(append([]string{}, path...), name), depth+1)
		}
	}
	walk(doc.Body, nil, 0)
	return out, nil
}

// Encode implements usecase.OPMLCodec: feeds grouped into one folder per
// category, feeds without a category at the top level.
func (OPML) Encode(feeds []domain.Feed, created time.Time) ([]byte, error) {
	doc := opmlDoc{Version: "2.0", Head: opmlHead{Title: "trisse-home feeds", DateCreated: created.UTC().Format(time.RFC1123Z)}}
	folders := map[string]int{}
	for _, f := range feeds {
		o := opmlOutline{Text: f.Title, Title: f.Title, Type: "rss", XMLURL: f.URL, HTMLURL: f.SiteURL}
		if f.Category == "" {
			doc.Body = append(doc.Body, o)
			continue
		}
		i, ok := folders[f.Category]
		if !ok {
			doc.Body = append(doc.Body, opmlOutline{Text: f.Category, Title: f.Category})
			i = len(doc.Body) - 1
			folders[f.Category] = i
		}
		doc.Body[i].Children = append(doc.Body[i].Children, o)
	}
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode OPML: %w", err)
	}
	return append([]byte(xml.Header), append(out, '\n')...), nil
}
