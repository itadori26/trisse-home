// Package presenter maps articles to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/readlater/domain"
)

// ArticleDTO is a saved article. Content (sanitised HTML) is only set when
// a single article is requested.
type ArticleDTO struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Site           string    `json:"site"`
	Excerpt        string    `json:"excerpt"`
	Content        string    `json:"content,omitempty"`
	ReadingMinutes int       `json:"readingMinutes"`
	Status         string    `json:"status" tstype:"'unread' | 'read' | 'archived'"`
	SavedAt        time.Time `json:"savedAt"`
	Bytes          int64     `json:"bytes"`
	// Error says why there is no reader copy (the link is kept).
	Error string `json:"error"`
	// Pending: imported, the reader copy is made in the background.
	Pending bool `json:"pending"`
}

// SettingsDTO are the read-later settings.
type SettingsDTO struct {
	// ArchiveDays deletes archived articles after so many days; 0 = keep.
	ArchiveDays int `json:"archiveDays"`
}

// ImportReportDTO says what an import did.
type ImportReportDTO struct {
	Added      int `json:"added"`
	Duplicates int `json:"duplicates"`
	Invalid    int `json:"invalid"`
}

// Article maps an article.
func Article(a domain.Article) ArticleDTO {
	return ArticleDTO{ID: a.ID, URL: a.URL, Title: a.Title, Site: a.Site, Excerpt: a.Excerpt, Content: a.Content,
		ReadingMinutes: a.ReadingMinutes(), Status: string(a.Status), SavedAt: a.SavedAt, Bytes: a.Bytes, Error: a.Error, Pending: a.Pending}
}

// Articles maps a list.
func Articles(as []domain.Article) []ArticleDTO {
	out := make([]ArticleDTO, 0, len(as))
	for _, a := range as {
		out = append(out, Article(a))
	}
	return out
}
