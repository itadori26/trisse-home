package controller

import (
	"encoding/json"
	"time"
)

// DocumentDTO is the export file.
type DocumentDTO struct {
	Format     string                     `json:"format"`
	Version    int                        `json:"version"`
	ExportedAt time.Time                  `json:"exportedAt"`
	Sections   map[string]json.RawMessage `json:"sections"`
}

// ImportResultDTO is returned by POST /api/transfer/import.
type ImportResultDTO struct {
	Imported []string          `json:"imported"`
	Skipped  []string          `json:"skipped"`
	Failed   map[string]string `json:"failed"`
}
