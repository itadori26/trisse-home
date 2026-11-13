package gateway

import (
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/rss/domain"
)

// A typical export of another reader: nested folders, feeds at the top
// level, title vs. text, a folder outline with type attributes.
const readerExport = `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Feedly export</title></head>
  <body>
    <outline text="Tech" title="Tech">
      <outline type="rss" text="Go Blog" title="The Go Blog" xmlUrl="https://go.dev/blog/feed.atom" htmlUrl="https://go.dev/blog"/>
      <outline text="Linux">
        <outline type="rss" text="LWN" xmlUrl="https://lwn.net/headlines/rss"/>
      </outline>
    </outline>
    <outline type="rss" text="tagesschau" xmlUrl="https://www.tagesschau.de/xml/rss2/"/>
    <outline text="Empty folder"/>
  </body>
</opml>`

func TestOPMLDecode(t *testing.T) {
	got, err := OPML{}.Decode([]byte(readerExport))
	if err != nil || len(got) != 3 {
		t.Fatalf("%+v %v", got, err)
	}
	want := []domain.ImportEntry{
		{Title: "The Go Blog", URL: "https://go.dev/blog/feed.atom", SiteURL: "https://go.dev/blog", Category: "Tech"},
		{Title: "LWN", URL: "https://lwn.net/headlines/rss", Category: "Tech / Linux"},
		{Title: "tagesschau", URL: "https://www.tagesschau.de/xml/rss2/"},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%d: %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestOPMLRefusesHostileXML(t *testing.T) {
	for name, doc := range map[string]string{
		"external entity": `<?xml version="1.0"?><!DOCTYPE opml [<!ENTITY x SYSTEM "file:///etc/passwd">]>
			<opml version="2.0"><body><outline text="&x;" xmlUrl="https://a.b/feed"/></body></opml>`,
		"entity expansion": `<?xml version="1.0"?><!DOCTYPE opml [<!ENTITY a "aaaaaaaaaa"><!ENTITY b "&a;&a;&a;&a;&a;">]>
			<opml version="2.0"><body><outline text="&b;" xmlUrl="https://a.b/feed"/></body></opml>`,
		"not opml": `<rss version="2.0"><channel><title>x</title></channel></rss>`,
		"broken":   `<opml><body><outline`,
	} {
		if entries, err := (OPML{}).Decode([]byte(doc)); err == nil {
			t.Errorf("%s: accepted %+v", name, entries)
		}
	}
	// Nesting deeper than maxDepth is cut off instead of recursing forever.
	deep := strings.Repeat(`<outline text="f">`, 50) + `<outline text="x" xmlUrl="https://a.b/f"/>` + strings.Repeat(`</outline>`, 50)
	if got, err := (OPML{}).Decode([]byte(`<opml version="2.0"><body>` + deep + `</body></opml>`)); err != nil || len(got) != 0 {
		t.Fatalf("deep: %+v %v", got, err)
	}
}

func TestOPMLRoundTrip(t *testing.T) {
	feeds := []domain.Feed{
		{Title: "Go & more", URL: "https://go.dev/blog/feed.atom?a=1&b=2", SiteURL: "https://go.dev/blog", Category: "Tech"},
		{Title: "tagesschau", URL: "https://www.tagesschau.de/xml/rss2/"},
		{Title: "LWN", URL: "https://lwn.net/headlines/rss", Category: "Tech"},
	}
	data, err := OPML{}.Encode(feeds, time.Date(2026, 11, 13, 22, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.HasPrefix(s, "<?xml") || !strings.Contains(s, `<opml version="2.0">`) || strings.Count(s, `text="Tech"`) != 1 {
		t.Fatalf("%s", s)
	}
	back, err := OPML{}.Decode(data)
	if err != nil || len(back) != 3 || back[0].Category != "Tech" || back[0].URL != feeds[0].URL || back[0].Title != "Go & more" || back[2].Category != "" {
		t.Fatalf("%+v %v\n%s", back, err, s)
	}
}
