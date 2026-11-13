// Package presenter maps new-releases use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/modules/releases/usecase"
)

// ArtistDTO is a MusicBrainz artist.
type ArtistDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
}

// ReleaseDTO is a release group; the cover is at /api/releases/covers/{id}.
type ReleaseDTO struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Type   string `json:"type"`
	// Date is YYYY-MM-DD.
	Date string `json:"date"`
}

// ListDTO is returned by GET /api/releases.
type ListDTO struct {
	Releases  []ReleaseDTO `json:"releases"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Error     string       `json:"error"`
}

// Artists maps artists.
func Artists(list []domain.Artist) []ArtistDTO {
	out := make([]ArtistDTO, 0, len(list))
	for _, a := range list {
		out = append(out, ArtistDTO(a))
	}
	return out
}

// List maps the release list.
func List(v usecase.ReleasesView) ListDTO {
	dto := ListDTO{UpdatedAt: v.UpdatedAt, Error: v.Error, Releases: make([]ReleaseDTO, 0, len(v.Releases))}
	for _, r := range v.Releases {
		dto.Releases = append(dto.Releases, ReleaseDTO{ID: r.ID, Title: r.Title, Artist: r.Artist, Type: r.Type, Date: r.Date.Format(time.DateOnly)})
	}
	return dto
}
