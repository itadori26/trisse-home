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

	"trisse-home/internal/modules/bookmarks/domain"
)

// Size limits of outbound reads (#72).
const (
	maxPageBytes = 512 << 10
	maxDrain     = 64 << 10
)

// Web implements usecase.MetadataFetcher and usecase.LinkChecker over
// HTTP. The client is expected to carry timeouts and the platform's
// internal address guard.
type Web struct{ client *http.Client }

// NewWeb returns the HTTP adapter.
func NewWeb(client *http.Client) *Web { return &Web{client: client} }

func (w *Web) get(ctx context.Context, method, rawURL, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", accept)
	return w.client.Do(req)
}

func drain(res *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, maxDrain))
	_ = res.Body.Close()
}

// Fetch implements usecase.MetadataFetcher: title and declared icon from
// the <head> of an HTML page (at most 512 KB are read).
func (w *Web) Fetch(ctx context.Context, pageURL string) (domain.PageMetadata, error) {
	res, err := w.get(ctx, http.MethodGet, pageURL, "text/html,application/xhtml+xml;q=0.9,*/*;q=0.1")
	if err != nil {
		return domain.PageMetadata{}, err
	}
	defer drain(res)
	if res.StatusCode >= http.StatusBadRequest {
		return domain.PageMetadata{}, fmt.Errorf("page answered %s", res.Status)
	}
	if mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); mt != "text/html" && mt != "application/xhtml+xml" {
		return domain.PageMetadata{}, errors.New("not an HTML page")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxPageBytes))
	if err != nil {
		return domain.PageMetadata{}, fmt.Errorf("read page: %w", err)
	}
	meta := ParseHead(string(body))
	// Relative icons resolve against the final URL after redirects.
	if meta.IconURL != "" && res.Request != nil && res.Request.URL != nil {
		meta.IconURL = domain.ResolveIconURL(res.Request.URL.String(), meta.IconURL)
	}
	return meta, nil
}

var (
	titlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title\s*>`)
	linkPattern  = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	attrPattern  = regexp.MustCompile(`(?is)([a-z:-]+)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+))`)
)

// ParseHead extracts the title and the best raster icon from HTML. Only
// the <head> is looked at; SVG icons are skipped.
func ParseHead(doc string) domain.PageMetadata {
	if i := strings.Index(strings.ToLower(doc), "</head>"); i >= 0 {
		doc = doc[:i]
	}
	var meta domain.PageMetadata
	if m := titlePattern.FindStringSubmatch(doc); m != nil {
		meta.Title = html.UnescapeString(m[1])
	}
	best := 99
	for _, tag := range linkPattern.FindAllString(doc, -1) {
		attrs := map[string]string{}
		for _, a := range attrPattern.FindAllStringSubmatch(tag, -1) {
			attrs[strings.ToLower(a[1])] = html.UnescapeString(a[2] + a[3] + a[4])
		}
		href := strings.TrimSpace(attrs["href"])
		if href == "" || strings.Contains(strings.ToLower(attrs["type"]), "svg") ||
			strings.HasSuffix(strings.ToLower(strings.SplitN(href, "?", 2)[0]), ".svg") {
			continue
		}
		rank := 99
		switch rel := " " + strings.ToLower(attrs["rel"]) + " "; {
		case strings.Contains(rel, " shortcut icon "):
			rank = 1
		case strings.Contains(rel, " icon "):
			rank = 0
		case strings.Contains(rel, " apple-touch-icon "):
			rank = 2
		}
		if rank < best {
			best, meta.IconURL = rank, href
		}
	}
	return meta
}

// FetchIcon implements usecase.MetadataFetcher. The type is taken from the
// content itself when it is a recognisable image, so a wrong header cannot
// smuggle in another format.
func (w *Web) FetchIcon(ctx context.Context, iconURL string) (domain.Icon, error) {
	res, err := w.get(ctx, http.MethodGet, iconURL, "image/*")
	if err != nil {
		return domain.Icon{}, err
	}
	defer drain(res)
	if res.StatusCode != http.StatusOK {
		return domain.Icon{}, fmt.Errorf("icon answered %s", res.Status)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, domain.MaxIconBytes+1))
	if err != nil {
		return domain.Icon{}, fmt.Errorf("read icon: %w", err)
	}
	ct := http.DetectContentType(data)
	if !strings.HasPrefix(ct, "image/") {
		ct = res.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(ct), "svg") || !strings.HasPrefix(strings.ToLower(ct), "image/") {
			return domain.Icon{}, errors.New("not a raster image")
		}
	}
	return domain.NewIcon(ct, data)
}

// Check implements usecase.LinkChecker: HEAD first, GET when the server
// does not support HEAD.
func (w *Web) Check(ctx context.Context, rawURL string) (int, string, bool) {
	code, final, retry, failed := w.request(ctx, http.MethodHead, rawURL)
	if retry {
		code, final, _, failed = w.request(ctx, http.MethodGet, rawURL)
	}
	return code, final, failed
}

// request performs one check request; retry reports answers to HEAD that
// some servers give although GET works.
func (w *Web) request(ctx context.Context, method, rawURL string) (code int, final string, retry, failed bool) {
	res, err := w.get(ctx, method, rawURL, "*/*")
	if err != nil {
		return 0, "", false, true
	}
	defer drain(res)
	if res.Request != nil && res.Request.URL != nil {
		final = res.Request.URL.String()
	}
	switch res.StatusCode {
	case http.StatusMethodNotAllowed, http.StatusNotImplemented, http.StatusForbidden, http.StatusNotFound:
		retry = method == http.MethodHead
	}
	return res.StatusCode, final, retry, false
}
