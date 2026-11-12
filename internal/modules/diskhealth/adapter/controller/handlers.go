// Package controller contains the disk health HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/diskhealth/adapter/presenter"
	"trisse-home/internal/modules/diskhealth/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the disk health endpoints.
type Handlers struct{ uc usecase.DiskHealth }

// NewHandlers returns disk health handlers.
func NewHandlers(uc usecase.DiskHealth) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/diskhealth.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	disks, err := h.uc.Disks(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Disks(disks))
}
