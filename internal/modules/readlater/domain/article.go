package domain

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Status of an article.
type Status string

// Statuses.
const (
	StatusUnread   Status = "unread"
	StatusRead     Status = "read"
	StatusArchived Status = "archived"
)

// ParseStatus returns the status or false.
func ParseStatus(s string) (Status, bool) {
	switch Status(s) {
	case StatusUnread, StatusRead, StatusArchived:
		return Status(s), true
	}
	return "", false
}

// Storage limits (#99).
const (
	// MaxContentBytes bounds the stored reader copy (HTML).
	MaxContentBytes = 1 << 20
	// MaxImages and MaxImageBytes bound the images kept per article.
	MaxImages     = 20
	MaxImageBytes = 2 << 20
	// MaxArticleImageBytes bounds all images of one article.
	MaxArticleImageBytes = 8 << 20
	// MaxStorageBytes bounds all stored copies; saving beyond it is refused
	// until archived articles are deleted.
	MaxStorageBytes = 200 << 20
	maxTitleRunes   = 300
	maxExcerptRunes = 300
	// WordsPerMinute is the reading speed of the reading-time estimate.
	WordsPerMinute = 220
)

// Errors.
var (
	ErrNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The article does not exist.")
	ErrStorageFull = kernel.NewPublicError(kernel.ErrConflict, "The read-later storage is full (200 MB). Delete archived articles first.")
)

// Article is a saved page.
type Article struct {
	ID      string
	URL     string
	Title   string
	Site    string
	Excerpt string
	// Content is the sanitised reader copy (HTML); empty when extraction
	// failed (the link is still kept).
	Content string
	Words   int
	Status  Status
	SavedAt time.Time
	// Bytes is the stored size of content and images.
	Bytes int64
	// Error says why no reader copy could be made.
	Error string
	// Pending articles were imported; the reader copy is fetched in the
	// background (#99).
	Pending bool
	// StatusAt is when the status last changed (zero: when saved).
	StatusAt time.Time
}

// Changed returns when the status last changed.
func (a Article) Changed() time.Time {
	if a.StatusAt.IsZero() {
		return a.SavedAt
	}
	return a.StatusAt
}

// Settings of the read-later list.
type Settings struct {
	// ArchiveDays deletes archived articles after so many days; 0 = keep.
	ArchiveDays int
}

// NewSettings validates the settings.
func NewSettings(archiveDays int) (Settings, error) {
	if archiveDays < 0 || archiveDays > 365 {
		return Settings{}, kernel.NewValidationError("archiveDays", "Archived articles can be kept for 1 to 365 days, or forever (0).")
	}
	return Settings{ArchiveDays: archiveDays}, nil
}

// ImportEntry is one link of an export file (Pocket CSV).
type ImportEntry struct {
	URL      string
	Title    string
	Added    time.Time
	Archived bool
}

// MaxImportEntries bounds one import.
const MaxImportEntries = 2000

// NewImported creates a pending article from an export entry; its reader
// copy is fetched later.
func NewImported(id string, e ImportEntry, now time.Time) (Article, error) {
	a, err := NewArticle(id, e.URL, now)
	if err != nil {
		return Article{}, err
	}
	if t := truncate(strings.Join(strings.Fields(e.Title), " "), maxTitleRunes); t != "" {
		a.Title = t
	}
	if !e.Added.IsZero() && e.Added.Before(now) {
		a.SavedAt = e.Added
	}
	if e.Archived {
		a.Status, a.StatusAt = StatusArchived, now
	}
	a.Pending = true
	return a, nil
}

// NewArticle validates the address of a page to save.
func NewArticle(id, rawURL string, now time.Time) (Article, error) {
	raw := strings.TrimSpace(rawURL)
	if i := strings.IndexByte(raw, '#'); i >= 0 {
		raw = raw[:i] // the fragment is not part of the page
	}
	// The central check of addresses the server requests (#191): http(s),
	// no credentials, never link-local or cloud metadata addresses.
	u, err := kernel.ParseTarget("url", raw, kernel.WithTargetQuery())
	if err != nil {
		return Article{}, err
	}
	return Article{ID: id, URL: u.String(), Title: u.Host, Site: u.Hostname(), Status: StatusUnread, SavedAt: now}, nil
}

// ApplyPage takes over what the extractor found.
func (a *Article) ApplyPage(title, site, excerpt, content string, words int) {
	if t := truncate(strings.Join(strings.Fields(title), " "), maxTitleRunes); t != "" {
		a.Title = t
	}
	if site = strings.TrimSpace(site); site != "" {
		a.Site = truncate(site, 100)
	}
	a.Excerpt = truncate(strings.Join(strings.Fields(excerpt), " "), maxExcerptRunes)
	a.Content, a.Words, a.Error = content, words, ""
}

// ReadingMinutes estimates the reading time, at least one minute.
func (a Article) ReadingMinutes() int {
	return max(1, (a.Words+WordsPerMinute-1)/WordsPerMinute)
}

// CountWords counts the words of plain text.
func CountWords(text string) int { return len(strings.Fields(text)) }

// imgPlaceholder matches the images of a sanitised copy before they are
// stored: <img src="cid:N" ...> (written only by the sanitiser).
var imgPlaceholder = regexp.MustCompile(`<img src="cid:(\d+)"([^>]*)>`)

// ResolveImages points the placeholders of stored images to their address
// and removes the images that could not be stored.
func ResolveImages(content string, stored map[int]string) string {
	return imgPlaceholder.ReplaceAllStringFunc(content, func(m string) string {
		sub := imgPlaceholder.FindStringSubmatch(m)
		n, _ := strconv.Atoi(sub[1])
		if src, ok := stored[n]; ok {
			return `<img src="` + src + `"` + sub[2] + `>`
		}
		return ""
	})
}

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n-1])) + "…"
}
