package gateway

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
)

// Article implements usecase.ArticleFetcher: it loads a feed item's page
// through the guarded client (no internal addresses) and keeps the readable
// text. A crude but dependency-free extraction: the <article> or <main>
// element if there is one, without scripts, styles and navigation.
type Article struct{ client *http.Client }

// NewArticle returns the fetcher.
func NewArticle(client *http.Client) *Article { return &Article{client: client} }

const maxArticleBytes = 2 << 20

var (
	dropBlocks = regexp.MustCompile(`(?is)<(script|style|noscript|nav|header|footer|aside|form|svg|figure)\b.*?</(?:script|style|noscript|nav|header|footer|aside|form|svg|figure)>|<!--.*?-->`)
	articleRe  = regexp.MustCompile(`(?is)<article\b[^>]*>(.*)</article>`)
	mainRe     = regexp.MustCompile(`(?is)<main\b[^>]*>(.*)</main>`)
	bodyRe     = regexp.MustCompile(`(?is)<body\b[^>]*>(.*)</body>`)
	blockTags  = regexp.MustCompile(`(?i)</?(?:p|div|br|li|h[1-6]|tr|section)\b[^>]*>`)
	anyTag     = regexp.MustCompile(`(?s)<[^>]*>`)
)

// Text implements usecase.ArticleFetcher.
func (a *Article) Text(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/html")
	res, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the page answered %s", res.Status)
	}
	if mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); mt != "text/html" && mt != "application/xhtml+xml" {
		return "", errors.New("the link is not a web page")
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxArticleBytes))
	if err != nil {
		return "", fmt.Errorf("read page: %w", err)
	}
	return ExtractText(string(raw)), nil
}

// ExtractText reduces an HTML page to its readable text.
func ExtractText(page string) string {
	page = dropBlocks.ReplaceAllString(page, " ")
	for _, re := range []*regexp.Regexp{articleRe, mainRe, bodyRe} {
		if m := re.FindStringSubmatch(page); m != nil {
			page = m[1]
			break
		}
	}
	page = blockTags.ReplaceAllString(page, "\n")
	page = html.UnescapeString(anyTag.ReplaceAllString(page, " "))
	lines := strings.Split(page, "\n")
	out := lines[:0]
	for _, l := range lines {
		if l = strings.Join(strings.Fields(l), " "); l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
