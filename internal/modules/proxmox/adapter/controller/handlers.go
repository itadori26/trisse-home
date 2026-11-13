// Package controller contains the Proxmox HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/proxmox/adapter/presenter"
	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/modules/proxmox/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the Proxmox endpoints.
type Handlers struct{ uc usecase.Proxmox }

// NewHandlers returns Proxmox handlers.
func NewHandlers(uc usecase.Proxmox) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/proxmox/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/proxmox/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Status handles GET /api/proxmox/status.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Status(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Status(res))
}

// Act handles POST /api/proxmox/guests/{id}/{action}.
func (h *Handlers) Act(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, domain.ErrGuestNotFound)
		return
	}
	if err := h.uc.Act(r.Context(), id, r.PathValue("action")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
