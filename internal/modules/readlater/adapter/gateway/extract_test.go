package gateway

import (
	"net/url"
	"os"
	"strings"
	"testing"
)

func mustURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}

func TestExtractArticle(t *testing.T) {
	doc, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}
	p := Extract(doc, mustURL("https://blog.example/2026/homeserver"))
	if p.Title != "Warum Homeserver glücklich machen" || p.Site != "Beispielblog" || p.Excerpt != "Ein Erfahrungsbericht nach drei Jahren." {
		t.Fatalf("meta: %+v", p)
	}
	for _, want := range []string{
		"<h2>Warum Homeserver glücklich machen</h2>",
		"<p>Seit drei Jahren",
		`<img src="cid:0" alt="Der Server im Flur">`,
		"<figcaption>Klein, leise, sparsam.</figcaption>",
		`<a href="https://example.org/wg" rel="noopener noreferrer nofollow" target="_blank">WireGuard</a>`,
	} {
		if !strings.Contains(p.Content, want) {
			t.Errorf("missing %q in\n%s", want, p.Content)
		}
	}
	for _, unwanted := range []string{"Cookies", "Start", "Beliebte", "Toller Artikel", "Teilen", "©", "track", "class=", "id="} {
		if strings.Contains(p.Content, unwanted) {
			t.Errorf("unwanted %q in\n%s", unwanted, p.Content)
		}
	}
	if len(p.Images) != 1 || p.Images[0] != "https://blog.example/img/server.jpg" || p.Words < 50 {
		t.Fatalf("images %v words %d", p.Images, p.Words)
	}
}

// TestSanitiserRefusesScriptVectors: whatever the page contains, the copy
// has no script, no event handler, no style and no non-http(s) link.
func TestSanitiserRefusesScriptVectors(t *testing.T) {
	body := `<article><p>` + strings.Repeat("Text, ", 60) + `</p>
		<script>alert(1)</script>
		<p onclick="alert(2)" style="background:url(javascript:alert(3))">click</p>
		<img src="x" onerror="alert(4)">
		<img src="javascript:alert(5)"><img src="data:image/svg+xml,<svg onload=alert(6)>">
		<a href="javascript:alert(7)">a</a><a href=" JaVaScRiPt:alert(8)">b</a><a href="&#106;avascript:alert(9)">c</a>
		<a href="vbscript:msgbox(10)">d</a><a href="data:text/html,<script>alert(11)</script>">e</a>
		<svg><script>alert(12)</script></svg><iframe src="https://evil.example"></iframe>
		<object data="x.swf"></object><embed src="x.swf"><form action="/x"><input value="y"></form>
		<math><mtext><img src=x onerror=alert(13)></mtext></math>
		<p>&lt;script&gt;alert(14)&lt;/script&gt;</p>
		<div><style>p{}</style><noscript><img src=x onerror=alert(15)></noscript></div>
		<a href="https://ok.example/?q=&quot;&gt;&lt;script&gt;">ok</a>
	</article>`
	p := Extract([]byte(body), mustURL("https://site.example/a"))
	low := strings.ToLower(p.Content)
	for _, bad := range []string{"<script", "onerror", "onclick", "onload", "style", "javascript:", "vbscript:", "data:", "<svg", "<iframe", "<object", "<embed", "<form", "<input", "<math", "<noscript"} {
		if strings.Contains(low, bad) {
			t.Errorf("%q survived:\n%s", bad, p.Content)
		}
	}
	// Escaped text stays text; quotes in URLs cannot break out of the attribute.
	if !strings.Contains(p.Content, "&lt;script&gt;alert(14)&lt;/script&gt;") {
		t.Errorf("text not escaped:\n%s", p.Content)
	}
	if !strings.Contains(p.Content, `<a href="https://ok.example/?q=%22%3E%3Cscript%3E"`) && !strings.Contains(p.Content, `href="https://ok.example/?q=&#34;&gt;&lt;script&gt;"`) {
		t.Errorf("link attribute:\n%s", p.Content)
	}
	if len(p.Images) != 1 || p.Images[0] != "https://site.example/x" {
		t.Fatalf("only the http(s) image is kept: %v", p.Images)
	}
}

func TestExtractWithoutArticle(t *testing.T) {
	p := Extract([]byte(`<html><head><title>Kurz</title></head><body><p>Nur ein Satz.</p></body></html>`), mustURL("https://x.example/"))
	if p.Title != "Kurz" || !strings.Contains(p.Content, "Nur ein Satz.") || p.Site != "x.example" {
		t.Fatalf("%+v", p)
	}
	if p := Extract([]byte(""), mustURL("https://x.example/")); p.Content != "" {
		t.Fatalf("empty: %+v", p)
	}
}
