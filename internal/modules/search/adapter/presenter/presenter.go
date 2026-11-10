// Package presenter maps search responses to JSON DTOs and renders the
// OpenSearch description document.
//
// Ring: Interface Adapters.
package presenter

import (
	"encoding/xml"

	"trisse-home/internal/modules/search/usecase"
)

// ProviderDTO describes a provider.
type ProviderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ShortcutDTO describes a custom shortcut and is the body of POST /api/search/shortcuts.
type ShortcutDTO struct {
	Keyword     string `json:"keyword"`
	Name        string `json:"name"`
	URLTemplate string `json:"urlTemplate"`
}

// SettingsDTO is returned by the search settings endpoints.
type SettingsDTO struct {
	DefaultProvider string        `json:"defaultProvider"`
	Providers       []ProviderDTO `json:"providers"`
	Shortcuts       []ShortcutDTO `json:"shortcuts"`
}

// Settings maps a SettingsResponse.
func Settings(r usecase.SettingsResponse) SettingsDTO {
	dto := SettingsDTO{DefaultProvider: r.DefaultProvider, Providers: []ProviderDTO{}, Shortcuts: []ShortcutDTO{}}
	for _, p := range r.Providers {
		dto.Providers = append(dto.Providers, ProviderDTO{ID: p.ID, Name: p.Name})
	}
	for _, s := range r.Shortcuts {
		dto.Shortcuts = append(dto.Shortcuts, ShortcutDTO{Keyword: s.Keyword, Name: s.Name, URLTemplate: s.URLTemplate})
	}
	return dto
}

type openSearchURL struct {
	Type     string `xml:"type,attr"`
	Method   string `xml:"method,attr"`
	Template string `xml:"template,attr"`
}

type openSearchDescription struct {
	XMLName       xml.Name      `xml:"OpenSearchDescription"`
	Xmlns         string        `xml:"xmlns,attr"`
	ShortName     string        `xml:"ShortName"`
	Description   string        `xml:"Description"`
	InputEncoding string        `xml:"InputEncoding"`
	URL           openSearchURL `xml:"Url"`
}

// OpenSearch renders the OpenSearch description that lets browsers use the
// start page as their search engine.
func OpenSearch(baseURL string) ([]byte, error) {
	doc := openSearchDescription{
		Xmlns: "http://a9.com/-/spec/opensearch/1.1/", ShortName: "trisse-home",
		Description: "Search via trisse-home", InputEncoding: "UTF-8",
		URL: openSearchURL{Type: "text/html", Method: "get", Template: baseURL + "/api/search?q={searchTerms}"},
	}
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}
