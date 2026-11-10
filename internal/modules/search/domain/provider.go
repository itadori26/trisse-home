// Package domain holds search providers (strategies), custom shortcuts and the
// query resolution rules.
//
// Ring: Entities.
package domain

import (
	"net/url"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// QueryPlaceholder marks where the encoded query is inserted into a URL template.
const QueryPlaceholder = "{query}"

// Provider is a search engine strategy that turns a query into a URL.
type Provider interface {
	ID() string
	Name() string
	URL(query string) string
}

// templateProvider is a Provider defined by a URL template.
type templateProvider struct {
	id, name, template string
}

func (p templateProvider) ID() string   { return p.id }
func (p templateProvider) Name() string { return p.name }
func (p templateProvider) URL(query string) string {
	return strings.ReplaceAll(p.template, QueryPlaceholder, url.QueryEscape(query))
}

// DefaultProviderID is DuckDuckGo, whose !bangs make shortcuts like !w or !g
// work without any configuration.
const DefaultProviderID = "duckduckgo"

var builtinProviders = []Provider{
	templateProvider{"duckduckgo", "DuckDuckGo", "https://duckduckgo.com/?q={query}"},
	templateProvider{"startpage", "Startpage", "https://www.startpage.com/do/search?q={query}"},
	templateProvider{"brave", "Brave Search", "https://search.brave.com/search?q={query}"},
	templateProvider{"ecosia", "Ecosia", "https://www.ecosia.org/search?q={query}"},
	templateProvider{"google", "Google", "https://www.google.com/search?q={query}"},
	templateProvider{"bing", "Bing", "https://www.bing.com/search?q={query}"},
}

// Providers returns the built-in providers.
func Providers() []Provider { return append([]Provider(nil), builtinProviders...) }

// ProviderByID returns the provider with the given ID.
func ProviderByID(id string) (Provider, error) {
	for _, p := range builtinProviders {
		if p.ID() == id {
			return p, nil
		}
	}
	return nil, kernel.NewValidationError("defaultProvider", "Unknown search provider.")
}

// ValidateTemplate checks that a URL template is an http(s) URL containing
// the query placeholder.
func ValidateTemplate(field, template string) error {
	if !strings.Contains(template, QueryPlaceholder) {
		return kernel.NewValidationError(field, "The URL must contain {query}.")
	}
	u, err := url.Parse(strings.ReplaceAll(template, QueryPlaceholder, "x"))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return kernel.NewValidationError(field, "The URL must start with http:// or https://.")
	}
	return nil
}
