package domain

import (
	"net/url"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// MaxQueryLength bounds the query to keep redirect URLs reasonable.
const MaxQueryLength = 1000

// Resolve returns the URL to open for query:
//  1. "!keyword rest" (or "rest !keyword") with a matching custom shortcut
//     uses that shortcut;
//  2. a query that is itself an http(s) URL is opened directly;
//  3. everything else goes to the default provider, which also handles the
//     provider's own bangs (e.g. DuckDuckGo's !w).
func Resolve(query string, provider Provider, shortcuts []Shortcut) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", kernel.NewValidationError("q", "The search query must not be empty.")
	}
	if len(query) > MaxQueryLength {
		return "", kernel.NewValidationError("q", "The search query is too long.")
	}
	if target, ok := matchShortcut(query, shortcuts); ok {
		return target, nil
	}
	if u, err := url.Parse(query); err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && !strings.Contains(query, " ") {
		return query, nil
	}
	return provider.URL(query), nil
}

func matchShortcut(query string, shortcuts []Shortcut) (string, bool) {
	fields := strings.Fields(query)
	for i, f := range fields {
		if !strings.HasPrefix(f, "!") || len(f) < 2 {
			continue
		}
		keyword := strings.ToLower(f[1:])
		for _, s := range shortcuts {
			if s.Keyword() == keyword {
				rest := strings.Join(append(append([]string{}, fields[:i]...), fields[i+1:]...), " ")
				return s.URL(rest), true
			}
		}
	}
	return "", false
}
