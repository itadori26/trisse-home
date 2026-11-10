package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

var keywordPattern = regexp.MustCompile(`^[a-z0-9]{1,16}$`)

const maxShortcutNameLength = 60

// Shortcut is a user-defined bang such as "!yt" → YouTube search.
type Shortcut struct {
	keyword  string
	name     string
	template string
}

// NewShortcut validates and creates a shortcut.
func NewShortcut(keyword, name, template string) (Shortcut, error) {
	keyword = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(keyword), "!"))
	if !keywordPattern.MatchString(keyword) {
		return Shortcut{}, kernel.NewValidationError("keyword", "The keyword must consist of 1 to 16 letters or digits.")
	}
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxShortcutNameLength {
		return Shortcut{}, kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	template = strings.TrimSpace(template)
	if err := ValidateTemplate("urlTemplate", template); err != nil {
		return Shortcut{}, err
	}
	return Shortcut{keyword: keyword, name: name, template: template}, nil
}

// Keyword returns the keyword without the leading "!".
func (s Shortcut) Keyword() string { return s.keyword }

// Name returns the display name.
func (s Shortcut) Name() string { return s.name }

// Template returns the URL template.
func (s Shortcut) Template() string { return s.template }

// URL builds the target URL for query.
func (s Shortcut) URL(query string) string {
	return templateProvider{template: s.template}.URL(query)
}
