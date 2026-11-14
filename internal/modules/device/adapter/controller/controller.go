// Package controller contains the device HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/device/adapter/presenter"
	"trisse-home/internal/modules/device/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves /api/device.
type Handlers struct{ uc usecase.Device }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Device) *Handlers { return &Handlers{uc: uc} }

// State handles GET /api/device.
func (h *Handlers) State(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.State(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.State(s))
}

// ChangeMode handles PUT /api/device/display-mode.
func (h *Handlers) ChangeMode(w http.ResponseWriter, r *http.Request) {
	var dto ModeDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.ChangeMode(r.Context(), dto.Mode)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.State(s))
}
