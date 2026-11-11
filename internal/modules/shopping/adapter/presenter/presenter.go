// Package presenter maps shopping list items to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/modules/shopping/usecase"
)

// SettingsDTO is returned by the settings endpoints (never the password).
type SettingsDTO struct {
	Configured  bool   `json:"configured"`
	ListURL     string `json:"listUrl"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
}

// ItemDTO is one list entry.
type ItemDTO struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
}

// ClearedDTO is returned by POST /api/shopping/clear-completed.
type ClearedDTO struct {
	Deleted int `json:"deleted"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Item maps one item.
func Item(it domain.Item) ItemDTO { return ItemDTO{ID: it.ID, Summary: it.Summary, Done: it.Done} }

// Items maps a list of items.
func Items(items []domain.Item) []ItemDTO {
	out := make([]ItemDTO, 0, len(items))
	for _, it := range items {
		out = append(out, Item(it))
	}
	return out
}
