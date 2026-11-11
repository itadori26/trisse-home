// Package controller contains the export/import HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// Handlers serves the transfer endpoints.
type Handlers struct{ uc usecase.Transfer }

// NewHandlers returns transfer handlers.
func NewHandlers(uc usecase.Transfer) *Handlers { return &Handlers{uc: uc} }

// Export handles GET /api/transfer/export.
func (h *Handlers) Export(w http.ResponseWriter, r *http.Request) {
	doc, err := h.uc.Export(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	dto := DocumentDTO{Format: doc.Format, Version: doc.Version, ExportedAt: doc.ExportedAt, Sections: map[string]json.RawMessage{}}
	for name, data := range doc.Sections {
		dto.Sections[name] = data
	}
	w.Header().Set("Content-Disposition", `attachment; filename="trisse-home-export-`+doc.ExportedAt.Format("20060102-150405")+`.json"`)
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, dto)
}

// Import handles POST /api/transfer/import with an export file as body.
func (h *Handlers) Import(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxImportBytes)
	var dto DocumentDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("file", "The file is larger than 16 MB."))
			return
		}
		httpapi.WriteDomainError(w, r, kernel.NewValidationError("file", "The file is not valid JSON."))
		return
	}
	doc := usecase.Document{Format: dto.Format, Version: dto.Version, ExportedAt: dto.ExportedAt, Sections: map[string][]byte{}}
	for name, raw := range dto.Sections {
		doc.Sections[name] = raw
	}
	res, err := h.uc.Import(r.Context(), doc)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ImportResultDTO(res))
}
