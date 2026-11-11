package domain

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

// MaxImport limits the number of bookmarks imported at once.
const MaxImport = 5000

// ImportCandidate is a bookmark read from a browser export.
type ImportCandidate struct {
	FolderPath  string
	Title       string
	URL         string
	Description string
	Tags        []string
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:max-1])) + "…"
}

// Importable reports whether the URL can be imported (http and https only;
// javascript:, place:, file: and the like are skipped).
func Importable(rawURL string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && len(rawURL) <= MaxURLLength
}

// SanitiseCandidate shortens over-long fields and drops tags that are not
// valid instead of rejecting the whole bookmark.
func SanitiseCandidate(c ImportCandidate) ImportCandidate {
	c.Title = truncate(c.Title, MaxTitleLength)
	c.Description = truncate(c.Description, MaxDescriptionLength)
	c.FolderPath = truncate(c.FolderPath, MaxFolderNameLength)
	var tags []string
	for _, t := range c.Tags {
		if norm, err := NormaliseTags([]string{t}); err == nil && len(norm) == 1 && len(tags) < MaxTags {
			tags = append(tags, norm[0])
		}
	}
	c.Tags = tags
	return c
}
