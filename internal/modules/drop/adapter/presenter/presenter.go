// Package presenter maps drop items to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/drop/domain"
)

// ItemDTO is a drop item; files are downloaded from /api/drop/{id}/file.
type ItemDTO struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind" tstype:"'text' | 'link' | 'file'"`
	Text        string    `json:"text"`
	FileName    string    `json:"fileName"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// Item maps an item.
func Item(i domain.Item) ItemDTO {
	return ItemDTO{ID: i.ID, Kind: string(i.Kind), Text: i.Text, FileName: i.FileName, ContentType: i.ContentType,
		Size: i.Size, CreatedAt: i.CreatedAt, ExpiresAt: i.ExpiresAt}
}

// Items maps a list.
func Items(is []domain.Item) []ItemDTO {
	out := make([]ItemDTO, 0, len(is))
	for _, i := range is {
		out = append(out, Item(i))
	}
	return out
}
