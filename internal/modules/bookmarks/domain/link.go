package domain

import (
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// LinkState is the result of the last link check (#72).
type LinkState string

// Link states; "" means not checked yet.
const (
	LinkUnchecked LinkState = ""
	LinkOK        LinkState = "ok"
	LinkRedirect  LinkState = "redirect"
	LinkBroken    LinkState = "broken"
)

// LinkCheckInterval is how often a link is checked again.
const LinkCheckInterval = 7 * 24 * time.Hour

// LinkStatus is a value object: state, HTTP status code, final URL of a
// redirect and the time of the check.
type LinkStatus struct {
	State     LinkState
	Code      int
	Target    string
	CheckedAt time.Time
}

// ClassifyLink turns the outcome of a request into a status. Services
// that answer 401/403 (login required) or 429 (rate limited) exist, so
// they count as reachable; a different final URL is a redirect.
func ClassifyLink(code int, original, final string, failed bool, now time.Time) LinkStatus {
	s := LinkStatus{Code: code, CheckedAt: now}
	switch {
	case failed || code == 0:
		s.State, s.Code = LinkBroken, 0
	case code >= 200 && code < 400, code == 401, code == 403, code == 429:
		s.State = LinkOK
		if final != "" && !sameURL(original, final) {
			s.State, s.Target = LinkRedirect, final
		}
	default:
		s.State = LinkBroken
	}
	return s
}

// sameURL ignores a trailing slash, the default port and the fragment.
func sameURL(a, b string) bool {
	norm := func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil {
			return raw
		}
		u.Fragment = ""
		u.Host = strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(u.Host), ":80"), ":443")
		u.Scheme = strings.ToLower(u.Scheme)
		u.Path = strings.TrimSuffix(u.Path, "/")
		return u.String()
	}
	return norm(a) == norm(b)
}

// Due reports whether the link should be checked again at now.
func (s LinkStatus) Due(now time.Time) bool {
	return s.CheckedAt.IsZero() || now.Sub(s.CheckedAt) >= LinkCheckInterval
}

// Settings switch the outbound features of the bookmarks module; both are
// off by default (#72).
type Settings struct {
	// Enrich fetches title and favicon when a bookmark is added.
	Enrich bool
	// CheckLinks checks every link once a week.
	CheckLinks bool
}

// PageMetadata is what a page tells about itself.
type PageMetadata struct {
	Title   string
	IconURL string
}

// CleanTitle collapses whitespace and shortens a page title to the limit.
func CleanTitle(s string) string {
	s = strings.Join(strings.FieldsFunc(s, unicode.IsSpace), " ")
	if utf8.RuneCountInString(s) > MaxTitleLength {
		s = string([]rune(s)[:MaxTitleLength-1]) + "…"
	}
	return s
}

// Icon is a stored favicon.
type Icon struct {
	ContentType string
	Data        []byte
}

// MaxIconBytes limits stored favicons.
const MaxIconBytes = 64 << 10

// iconTypes are the accepted raster formats; SVG is refused because it can
// carry scripts.
var iconTypes = []string{"image/png", "image/x-icon", "image/vnd.microsoft.icon", "image/jpeg", "image/gif", "image/webp"}

// ErrIconNotFound is returned for bookmarks without a stored favicon.
var ErrIconNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The bookmark has no icon.")

// NewIcon validates type and size of a favicon.
func NewIcon(contentType string, data []byte) (Icon, error) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if !slices.Contains(iconTypes, ct) || len(data) == 0 || len(data) > MaxIconBytes {
		return Icon{}, kernel.NewValidationError("icon", "The icon must be a PNG, ICO, JPEG, GIF or WebP image of at most 64 KB.")
	}
	return Icon{ContentType: ct, Data: data}, nil
}

// ResolveIconURL returns the absolute icon URL of a page: the declared
// icon relative to the page, or /favicon.ico of its origin.
func ResolveIconURL(pageURL, declared string) string {
	base, err := url.Parse(pageURL)
	if err != nil {
		return ""
	}
	if declared != "" {
		if ref, err := url.Parse(declared); err == nil {
			if u := base.ResolveReference(ref); u.Scheme == "http" || u.Scheme == "https" {
				return u.String()
			}
		}
	}
	return (&url.URL{Scheme: base.Scheme, Host: base.Host, Path: "/favicon.ico"}).String()
}
