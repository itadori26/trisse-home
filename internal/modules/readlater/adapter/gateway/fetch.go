package gateway

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/modules/readlater/usecase"
	"trisse-home/internal/shared/kernel"
)

// maxPageBytes bounds a fetched page.
const maxPageBytes = 5 << 20

// Fetcher implements usecase.PageReader and usecase.ImageFetcher with a
// guarded client (no loopback, link-local or metadata addresses, checked on
// every connection and redirect).
type Fetcher struct{ client *http.Client }

// NewFetcher returns a fetcher.
func NewFetcher(client *http.Client) *Fetcher { return &Fetcher{client: client} }

func (f *Fetcher) get(ctx context.Context, rawURL, accept string, limit int64) ([]byte, string, *url.URL, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", accept)
	res, err := f.client.Do(req)
	if err != nil {
		return nil, "", nil, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", nil, kernel.NewPublicErrorf(kernel.ErrInvalid, "The server answered {status}.", "status", res.Status)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, "", nil, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if int64(len(body)) > limit {
		return nil, "", nil, kernel.NewPublicError(kernel.ErrInvalid, "The page is too large.")
	}
	ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	return body, ct, res.Request.URL, nil
}

// Read implements usecase.PageReader.
func (f *Fetcher) Read(ctx context.Context, rawURL string) (usecase.Page, error) {
	body, ct, final, err := f.get(ctx, rawURL, "text/html,application/xhtml+xml", maxPageBytes)
	if err != nil {
		return usecase.Page{}, err
	}
	if ct != "" && ct != "text/html" && ct != "application/xhtml+xml" {
		return usecase.Page{}, kernel.NewPublicError(kernel.ErrInvalid, "The address is not a web page.")
	}
	p := Extract(body, final)
	return usecase.Page{Title: p.Title, Site: p.Site, Excerpt: p.Excerpt, Content: p.Content, Images: p.Images, Words: p.Words}, nil
}

// Fetch implements usecase.ImageFetcher: raster images only (no SVG, which
// can carry scripts), at most domain.MaxImageBytes.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (usecase.Image, error) {
	body, ct, _, err := f.get(ctx, rawURL, "image/avif,image/webp,image/png,image/jpeg,image/gif", domain.MaxImageBytes)
	if err != nil {
		return usecase.Image{}, err
	}
	switch ct {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/avif":
	default:
		// Trust the bytes, not the header, for the common formats.
		if sniffed := http.DetectContentType(body); strings.HasPrefix(sniffed, "image/") && sniffed != "image/svg+xml" {
			ct = sniffed
		} else {
			return usecase.Image{}, fmt.Errorf("not a raster image: %q", ct)
		}
	}
	return usecase.Image{Bytes: body, ContentType: ct}, nil
}
