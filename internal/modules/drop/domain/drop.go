package domain

import (
	"net/url"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Kind of an item.
type Kind string

// Kinds.
const (
	KindText Kind = "text"
	KindLink Kind = "link"
	KindFile Kind = "file"
)

// Limits (#101).
const (
	MaxTextBytes  = 100 << 10
	MaxFileBytes  = 25 << 20
	QuotaBytes    = 200 << 20
	MaxItems      = 100
	DefaultExpiry = 24 * time.Hour
	maxNameRunes  = 120
)

// Expiries offered to the user.
var Expiries = []time.Duration{time.Hour, DefaultExpiry, 7 * 24 * time.Hour}

// Errors.
var (
	ErrNotFound  = kernel.NewPublicError(kernel.ErrNotFound, "The item does not exist or has expired.")
	ErrQuotaFull = kernel.NewPublicError(kernel.ErrConflict, "The drop zone is full (200 MB, 100 items). Delete items first.")
	errEmpty     = kernel.NewValidationError("text", "Enter some text or a link.")
	errTextLarge = kernel.NewValidationError("text", "The text must be at most 100 KB.")
	errFileLarge = kernel.NewValidationError("file", "The file must be at most 25 MB.")
	errBadExpiry = kernel.NewValidationError("expiresIn", "Choose 1 hour, 24 hours or 7 days.")
	errEmptyFile = kernel.NewValidationError("file", "The file is empty.")
)

// Item is something dropped on one device to pick up on another.
type Item struct {
	ID   string
	Kind Kind
	// Text is the text or link; empty for files.
	Text string
	// FileName, ContentType and Size describe a file; Size is also the text size.
	FileName    string
	ContentType string
	Size        int64
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// Expired reports whether the item is gone at now.
func (i Item) Expired(now time.Time) bool { return !now.Before(i.ExpiresAt) }

func expiry(d time.Duration) (time.Duration, error) {
	if d == 0 {
		return DefaultExpiry, nil
	}
	for _, e := range Expiries {
		if d == e {
			return d, nil
		}
	}
	return 0, errBadExpiry
}

// NewText creates a text item; a single http(s) address becomes a link.
func NewText(id, text string, expiresIn time.Duration, now time.Time) (Item, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Item{}, errEmpty
	}
	if len(text) > MaxTextBytes {
		return Item{}, errTextLarge
	}
	d, err := expiry(expiresIn)
	if err != nil {
		return Item{}, err
	}
	kind := KindText
	if u, err := url.Parse(text); err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && !strings.ContainsAny(text, " \n\t") {
		kind = KindLink
	}
	return Item{ID: id, Kind: kind, Text: text, Size: int64(len(text)), CreatedAt: now, ExpiresAt: now.Add(d)}, nil
}

// NewFile creates a file item; the name is reduced to a safe base name.
func NewFile(id, name, contentType string, size int64, expiresIn time.Duration, now time.Time) (Item, error) {
	if size <= 0 {
		return Item{}, errEmptyFile
	}
	if size > MaxFileBytes {
		return Item{}, errFileLarge
	}
	d, err := expiry(expiresIn)
	if err != nil {
		return Item{}, err
	}
	return Item{ID: id, Kind: KindFile, FileName: CleanFileName(name), ContentType: contentType, Size: size, CreatedAt: now, ExpiresAt: now.Add(d)}, nil
}

// CleanFileName keeps the base name without path, control characters or
// quotes, at most 120 characters; "file" when nothing is left.
func CleanFileName(name string) string {
	name = path.Base(strings.ReplaceAll(strings.TrimSpace(name), `\`, "/"))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' || r == '/' {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	if utf8.RuneCountInString(name) > maxNameRunes {
		r := []rune(name)
		name = string(r[:maxNameRunes])
	}
	return name
}

// CheckQuota refuses an item that would exceed the item count or the total size.
func CheckQuota(items []Item, now time.Time, size int64) error {
	var used int64
	n := 0
	for _, i := range items {
		if !i.Expired(now) {
			used += i.Size
			n++
		}
	}
	if n >= MaxItems || used+size > QuotaBytes {
		return ErrQuotaFull
	}
	return nil
}

// SafeInlineType is the content type a download is served with: well-known
// passive types keep theirs, everything else (HTML, SVG, scripts) is sent as
// application/octet-stream.
func SafeInlineType(ct string) string {
	switch ct {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "application/pdf", "text/plain":
		return ct
	}
	return "application/octet-stream"
}
