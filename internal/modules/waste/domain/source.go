package domain

import (
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// SourceKind says where the collection dates come from (#153).
type SourceKind string

// Source kinds; the zero value means no source is configured.
const (
	SourceNone SourceKind = ""
	SourceURL  SourceKind = "url"
	SourceFile SourceKind = "file"
)

// Limits of an uploaded calendar.
const (
	// MaxCalendarBytes is the largest ICS file that can be uploaded.
	MaxCalendarBytes = 1 << 20
	maxFileNameRunes = 100
	// ExpiryWarning is how long before its last date an uploaded calendar
	// asks for next year's file.
	ExpiryWarning = 30 * 24 * time.Hour
)

// CalendarFile describes an uploaded ICS file; the content is stored apart.
type CalendarFile struct {
	Name       string
	UploadedAt time.Time
	// Events is the number of collection dates found in the file.
	Events int
	// From and To are the first and last collection date.
	From, To time.Time
}

// Source is the origin of the collection dates: an ICS URL or an uploaded file.
type Source struct {
	Kind SourceKind
	URL  string
	File CalendarFile
}

// NewURLSource validates an ICS subscription URL; webcal:// becomes https://.
func NewURLSource(rawURL string) (Source, error) {
	u, err := normaliseURL(rawURL)
	if err != nil {
		return Source{}, err
	}
	return Source{Kind: SourceURL, URL: u}, nil
}

// NewFileSource describes an uploaded file from the dates parsed out of it;
// a file without any date is rejected.
func NewFileSource(name string, uploadedAt time.Time, cs []Collection) (Source, error) {
	if len(cs) == 0 {
		return Source{}, kernel.NewValidationError("file", "The file contains no collection dates.")
	}
	f := CalendarFile{Name: cleanFileName(name), UploadedAt: uploadedAt.UTC(), Events: len(cs), From: cs[0].Date, To: cs[0].Date}
	for _, c := range cs[1:] {
		if c.Date.Before(f.From) {
			f.From = c.Date
		}
		if c.Date.After(f.To) {
			f.To = c.Date
		}
	}
	return Source{Kind: SourceFile, File: f}, nil
}

// Configured reports whether there is a source.
func (s Source) Configured() bool { return s.Kind != SourceNone }

// ExpiresWithin reports whether an uploaded calendar has its last date
// within d from now (or already behind it); subscriptions never expire.
func (s Source) ExpiresWithin(now time.Time, d time.Duration) bool {
	return s.Kind == SourceFile && !Date(now.Add(d)).Before(s.File.To)
}

// cleanFileName keeps the base name of an uploaded file, without control
// characters and at most 100 characters long.
func cleanFileName(name string) string {
	name = path.Base(strings.ReplaceAll(strings.TrimSpace(name), `\`, "/"))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)
	if name == "." || name == "/" || name == "" {
		name = "calendar.ics"
	}
	if utf8.RuneCountInString(name) > maxFileNameRunes {
		name = string([]rune(name)[:maxFileNameRunes])
	}
	return name
}
