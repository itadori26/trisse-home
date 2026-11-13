package gateway

import (
	"strings"
	"testing"
	"time"
)

func mustParse(t *testing.T, doc string) parsedFeed {
	t.Helper()
	f, err := parseFeed(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, doc)
	}
	return f
}

func TestParseRSS2(t *testing.T) {
	f := mustParse(t, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:dc="http://purl.org/dc/elements/1.1/">
 <channel>
  <title><![CDATA[Tech & News]]></title>
  <atom:link href="https://example.org/feed" rel="self" type="application/rss+xml"/>
  <link>https://example.org/</link>
  <atom:link href="https://example.org/feed2" rel="hub"/>
  <item>
   <title>One &amp; only</title>
   <itunes:title>Podcast title that must not win</itunes:title>
   <link>https://example.org/1</link>
   <guid isPermaLink="false">id-1</guid>
   <pubDate>Fri, 13 Nov 2026 10:30:00 EST</pubDate>
  </item>
  <item>
   <title>No guid</title>
   <link>https://example.org/2</link>
   <dc:date>2026-11-12T08:00:00+01:00</dc:date>
  </item>
  <item><title>Only a title&nbsp;here</title></item>
 </channel>
</rss>`)
	if f.Title != "Tech & News" || f.Link != "https://example.org/" || len(f.Items) != 3 {
		t.Fatalf("feed: %+v", f)
	}
	one, two, three := f.Items[0], f.Items[1], f.Items[2]
	if one.Title != "One & only" || one.ID != "id-1" || one.Link != "https://example.org/1" {
		t.Errorf("item 1: %+v", one)
	}
	if want := time.Date(2026, 11, 13, 15, 30, 0, 0, time.UTC); !one.Published.Equal(want) { // EST is UTC-5
		t.Errorf("pubDate with a zone abbreviation: %v, want %v", one.Published, want)
	}
	if two.ID != "https://example.org/2" || two.Published.IsZero() {
		t.Errorf("item 2 (guid falls back to the link, dc:date): %+v", two)
	}
	if !strings.HasPrefix(three.Title, "Only a title") {
		t.Errorf("item 3: %+v", three)
	}
}

func TestParseRSS1(t *testing.T) {
	f := mustParse(t, `<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/">
 <channel rdf:about="https://example.org/"><title>Old school</title><link>https://example.org/</link></channel>
 <item rdf:about="https://example.org/a"><title>A</title><link>https://example.org/a</link><dc:date>2026-11-01T10:00:00Z</dc:date></item>
 <item rdf:about="https://example.org/b"><title>B</title><link>https://example.org/b</link></item>
</rdf:RDF>`)
	if f.Title != "Old school" || len(f.Items) != 2 || f.Items[0].Published.IsZero() || f.Items[1].Link != "https://example.org/b" {
		t.Fatalf("%+v", f)
	}
}

func TestParseAtom(t *testing.T) {
	f := mustParse(t, `<?xml version="1.0" encoding="utf-8"?>
<a:feed xmlns:a="http://www.w3.org/2005/Atom" xmlns:foreign="urn:foreign">
 <a:title>Blog</a:title><foreign:title>Foreign</foreign:title>
 <a:link rel="self" href="https://blog.example/atom"/><a:link href="https://blog.example/"/>
 <a:entry><a:id>urn:1</a:id><a:title type="xhtml"><div xmlns="http://www.w3.org/1999/xhtml">Hello <b>world</b></div></a:title>
  <a:link rel="replies" href="https://blog.example/1/comments"/><a:link rel="alternate" href="https://blog.example/1"/>
  <a:published>2026-11-01T10:00:00Z</a:published><a:updated>2026-11-05T10:00:00Z</a:updated></a:entry>
 <a:entry><a:title>No id</a:title><a:link href="/relative"/><a:updated>2026-11-02T10:00:00Z</a:updated></a:entry>
</a:feed>`)
	if f.Title != "Blog" || f.Link != "https://blog.example/" || len(f.Items) != 2 {
		t.Fatalf("feed: %+v", f)
	}
	e := f.Items[0]
	if e.Title != "Hello world" || e.Link != "https://blog.example/1" || e.ID != "urn:1" || !e.Published.Equal(time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("entry 1 (xhtml title, alternate link, published before updated): %+v", e)
	}
	if n := f.Items[1]; n.Link != "https://blog.example/relative" || n.ID != n.Link {
		t.Errorf("entry 2 (relative link resolved against the site, id falls back to the link): %+v", n)
	}
}

func TestParseXMLBase(t *testing.T) {
	f := mustParse(t, `<rss version="2.0"><channel xml:base="https://example.org/feeds/"><title>t</title>
 <item><title>i</title><link>post/1</link></item></channel></rss>`)
	if got := f.Items[0].Link; got != "https://example.org/feeds/post/1" {
		t.Fatalf("link %q", got)
	}
}

func TestParseJSONFeed(t *testing.T) {
	f := mustParse(t, `{"version":"https://jsonfeed.org/version/1.1","title":"JSON","home_page_url":"https://j.example/",
	 "items":[{"id":"1","url":"https://j.example/1","title":"One","date_published":"2026-11-01T10:00:00Z"},
	          {"url":"https://j.example/2","date_modified":"2026-11-02T10:00:00+02:00"}]}`)
	if f.Title != "JSON" || f.Link != "https://j.example/" || len(f.Items) != 2 {
		t.Fatalf("%+v", f)
	}
	if f.Items[0].ID != "1" || f.Items[0].Published.IsZero() || f.Items[1].ID != "https://j.example/2" || f.Items[1].Published.IsZero() {
		t.Fatalf("items: %+v", f.Items)
	}
}

func TestParseCharsets(t *testing.T) {
	// "Äpfel & Birnen" as ISO-8859-1 (0xC4) and as windows-1252 with a typographic dash (0x96).
	latin1 := "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><rss><channel><title>\xc4pfel</title></channel></rss>"
	if f := mustParse(t, latin1); f.Title != "Äpfel" {
		t.Errorf("latin1: %q", f.Title)
	}
	cp1252 := "<?xml version=\"1.0\" encoding=\"windows-1252\"?><rss><channel><title>Nord\x96Süd</title></channel></rss>"
	cp1252 = strings.Replace(cp1252, "Süd", "S\xfcd", 1)
	if f := mustParse(t, cp1252); f.Title != "Nord–Süd" {
		t.Errorf("windows-1252: %q", f.Title)
	}
	if _, err := parseFeed(strings.NewReader(`<?xml version="1.0" encoding="x-unknown"?><rss><channel/></rss>`)); err == nil {
		t.Error("an unknown character set must be reported, not shown as garbage")
	}
}

func TestParseToleratesRealWorldDefects(t *testing.T) {
	t.Run("control characters", func(t *testing.T) {
		f := mustParse(t, "<rss><channel><title>A\x08B</title><item><title>x\x0b</title></item></channel></rss>")
		if f.Title != "AB" || len(f.Items) != 1 {
			t.Fatalf("%+v", f)
		}
	})
	t.Run("BOM and whitespace before the declaration", func(t *testing.T) {
		mustParse(t, "\xef\xbb\xbf\n  <?xml version=\"1.0\"?><rss><channel><title>x</title></channel></rss>")
	})
	t.Run("truncated download keeps what arrived", func(t *testing.T) {
		f := mustParse(t, "<rss><channel><title>T</title><item><title>one</title><link>https://e.org/1</link></item><item><title>tw")
		if f.Title != "T" || len(f.Items) == 0 || f.Items[0].Title != "one" {
			t.Fatalf("%+v", f)
		}
	})
	t.Run("a link before the item text is not auto-closed like in HTML", func(t *testing.T) {
		if f := mustParse(t, "<rss><channel><link>https://e.org/</link></channel></rss>"); f.Link != "https://e.org/" {
			t.Fatalf("%+v", f)
		}
	})
	t.Run("item without title or link is kept (the use case decides)", func(t *testing.T) {
		if f := mustParse(t, `<rss><channel><item><enclosure url="https://e.org/a.mp3"/></item></channel></rss>`); len(f.Items) != 1 {
			t.Fatalf("%+v", f)
		}
	})
}

func TestParseRejectsOtherDocuments(t *testing.T) {
	for name, doc := range map[string]string{
		"html page":     "<!doctype html><html><body><h1>Hi</h1></body></html>",
		"plain xml":     "<?xml version=\"1.0\"?><sitemap><url/></sitemap>",
		"text":          "hello",
		"empty":         "",
		"json, no feed": `{"a":1}`,
		"broken json":   `{"title":`,
	} {
		if _, err := parseFeed(strings.NewReader(doc)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseDate(t *testing.T) {
	utc := func(y int, m time.Month, d, h, mi, s int) time.Time { return time.Date(y, m, d, h, mi, s, 0, time.UTC) }
	for in, want := range map[string]time.Time{
		"Fri, 13 Nov 2026 10:30:00 GMT":      utc(2026, 11, 13, 10, 30, 0),
		"Fri, 13 Nov 2026 10:30:00 +0100":    utc(2026, 11, 13, 9, 30, 0),
		"Fri, 13 Nov 2026 10:30 +0000":       utc(2026, 11, 13, 10, 30, 0),
		"Fri,   3 Nov 2026 10:30:00 PST":     utc(2026, 11, 3, 18, 30, 0),
		"13 Nov 2026 10:30:00 CET":           utc(2026, 11, 13, 9, 30, 0),
		"2026-11-13T10:30:00Z":               utc(2026, 11, 13, 10, 30, 0),
		"2026-11-13T10:30:00.123+02:00":      time.Date(2026, 11, 13, 8, 30, 0, 123000000, time.UTC),
		"2026-11-13 10:30:00":                utc(2026, 11, 13, 10, 30, 0),
		"2026-11-13":                         utc(2026, 11, 13, 0, 0, 0),
		"  Fri, 13 Nov 2026 10:30:00 UT\n  ": utc(2026, 11, 13, 10, 30, 0),
	} {
		got, ok := parseDate(in)
		if !ok || !got.Equal(want) {
			t.Errorf("%q: got %v (%v), want %v", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "yesterday", "13/11/2026"} {
		if _, ok := parseDate(bad); ok {
			t.Errorf("%q should not parse", bad)
		}
	}
}
