package gateway

import (
	"bytes"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Page is what the extractor found on a web page (#99).
type Page struct {
	Title, Site, Excerpt string
	// Content is sanitised HTML; images are <img src="cid:N" alt="…"> with
	// N the index into Images.
	Content string
	Images  []string
	Words   int
}

// Readability-style extraction: take <article>/<main> or the block with the
// most paragraph text, then rebuild it from an allowlist of elements.

// dropped elements are removed with their content.
var dropped = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Noscript: true, atom.Iframe: true, atom.Object: true, atom.Embed: true,
	atom.Form: true, atom.Button: true, atom.Input: true, atom.Select: true, atom.Textarea: true, atom.Svg: true,
	atom.Math: true, atom.Nav: true, atom.Footer: true, atom.Aside: true, atom.Header: true, atom.Template: true,
	atom.Canvas: true, atom.Video: true, atom.Audio: true, atom.Link: true, atom.Meta: true, atom.Head: true,
}

// kept elements are copied (without attributes unless handled below);
// everything else is unwrapped (its children are kept).
var kept = map[atom.Atom]bool{
	atom.P: true, atom.Br: true, atom.Hr: true, atom.H2: true, atom.H3: true, atom.H4: true, atom.H5: true, atom.H6: true,
	atom.Ul: true, atom.Ol: true, atom.Li: true, atom.Blockquote: true, atom.Pre: true, atom.Code: true,
	atom.Em: true, atom.Strong: true, atom.B: true, atom.I: true, atom.U: true, atom.S: true, atom.Sub: true, atom.Sup: true,
	atom.Figure: true, atom.Figcaption: true, atom.Table: true, atom.Thead: true, atom.Tbody: true, atom.Tr: true,
	atom.Th: true, atom.Td: true, atom.Dl: true, atom.Dt: true, atom.Dd: true, atom.A: true, atom.Img: true,
}

var noise = regexp.MustCompile(`(?i)comment|footer|nav|sidebar|menu|share|social|related|promo|advert|\bads?\b|cookie|banner|newsletter|subscribe|popup`)

// Extract reads a page. base resolves relative links and images.
func Extract(doc []byte, base *url.URL) Page {
	root, err := html.Parse(bytes.NewReader(doc))
	if err != nil {
		return Page{}
	}
	p := Page{Title: title(root), Site: meta(root, "og:site_name")}
	if p.Site == "" {
		p.Site = base.Hostname()
	}
	p.Excerpt = meta(root, "og:description")
	if p.Excerpt == "" {
		p.Excerpt = meta(root, "description")
	}
	body := mainContent(root)
	if body == nil {
		return p
	}
	s := &sanitiser{base: base}
	var out bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		s.write(&out, c)
	}
	p.Content, p.Images, p.Words = strings.TrimSpace(out.String()), s.images, s.words
	if p.Excerpt == "" {
		p.Excerpt = s.firstText
	}
	return p
}

func find(n *html.Node, match func(*html.Node) bool) *html.Node {
	if match(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if f := find(c, match); f != nil {
			return f
		}
	}
	return nil
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func text(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && dropped[n.DataAtom] {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func meta(root *html.Node, name string) string {
	m := find(root, func(n *html.Node) bool {
		return n.DataAtom == atom.Meta && (attr(n, "property") == name || attr(n, "name") == name)
	})
	if m == nil {
		return ""
	}
	return strings.TrimSpace(attr(m, "content"))
}

func title(root *html.Node) string {
	if t := meta(root, "og:title"); t != "" {
		return t
	}
	if t := find(root, func(n *html.Node) bool { return n.DataAtom == atom.Title }); t != nil {
		return text(t)
	}
	if h := find(root, func(n *html.Node) bool { return n.DataAtom == atom.H1 }); h != nil {
		return text(h)
	}
	return ""
}

// mainContent picks the element that holds the article.
func mainContent(root *html.Node) *html.Node {
	if a := find(root, func(n *html.Node) bool { return n.DataAtom == atom.Article }); a != nil && len(text(a)) > 200 {
		return a
	}
	if m := find(root, func(n *html.Node) bool { return n.DataAtom == atom.Main || attr(n, "role") == "main" }); m != nil && len(text(m)) > 200 {
		return m
	}
	// The block whose direct paragraphs hold the most text wins.
	var best *html.Node
	bestScore := 0
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if dropped[n.DataAtom] {
				return
			}
			score := 0
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.DataAtom == atom.P {
					t := text(c)
					score += len(t) + 10*strings.Count(t, ",")
				}
			}
			if noise.MatchString(attr(n, "class") + " " + attr(n, "id")) {
				score /= 4
			}
			if score > bestScore {
				best, bestScore = n, score
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	if best != nil {
		return best
	}
	return find(root, func(n *html.Node) bool { return n.DataAtom == atom.Body })
}

// sanitiser writes an allowlisted copy: only kept elements, only the
// attributes below, every text escaped by the HTML renderer.
type sanitiser struct {
	base      *url.URL
	images    []string
	words     int
	firstText string
}

func (s *sanitiser) write(w *bytes.Buffer, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		if t := n.Data; strings.TrimSpace(t) != "" {
			s.words += len(strings.Fields(t))
			if s.firstText == "" {
				s.firstText = strings.Join(strings.Fields(t), " ")
			}
		}
		w.WriteString(html.EscapeString(n.Data))
		return
	case html.ElementNode:
	default:
		return // comments, doctype
	}
	if dropped[n.DataAtom] || noise.MatchString(attr(n, "class")+" "+attr(n, "id")) && n.DataAtom != atom.P {
		return
	}
	tag := n.DataAtom
	if tag == atom.H1 {
		tag = atom.H2 // the reader shows the title as the only h1
	}
	if !kept[tag] {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			s.write(w, c)
		}
		return
	}
	switch tag {
	case atom.Img:
		src := s.resolve(firstNonEmpty(attr(n, "src"), attr(n, "data-src")))
		if src == "" || len(s.images) >= maxImagesInPage {
			return
		}
		s.images = append(s.images, src)
		w.WriteString(`<img src="cid:` + strconv.Itoa(len(s.images)-1) + `" alt="` + html.EscapeString(attr(n, "alt")) + `">`)
		return
	case atom.Br, atom.Hr:
		w.WriteString("<" + tag.String() + ">")
		return
	case atom.A:
		href := s.resolve(attr(n, "href"))
		if href == "" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				s.write(w, c)
			}
			return
		}
		w.WriteString(`<a href="` + html.EscapeString(href) + `" rel="noopener noreferrer nofollow" target="_blank">`)
	default:
		w.WriteString("<" + tag.String() + ">")
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.write(w, c)
	}
	w.WriteString("</" + tag.String() + ">")
}

// maxImagesInPage bounds the image list before downloading.
const maxImagesInPage = 40

// resolve makes an absolute http(s) URL or returns "".
func (s *sanitiser) resolve(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := s.base.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return u.String()
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
