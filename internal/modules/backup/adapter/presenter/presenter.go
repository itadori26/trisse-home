// Package presenter maps backup files to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/backup/usecase"
)

// FileDTO is a backup in API responses.
type FileDTO struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// File maps a FileView.
func File(v usecase.FileView) FileDTO { return FileDTO(v) }

// Files maps file views.
func Files(vs []usecase.FileView) []FileDTO {
	out := make([]FileDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, File(v))
	}
	return out
}
