// Package controller contains the WireGuard HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/wireguard/adapter/presenter"
	"trisse-home/internal/modules/wireguard/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the WireGuard endpoints.
type Handlers struct{ uc usecase.WireGuard }

// NewHandlers returns WireGuard handlers.
func NewHandlers(uc usecase.WireGuard) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/wireguard/peers.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(res))
}

// Rename handles PUT /api/wireguard/names.
func (h *Handlers) Rename(w http.ResponseWriter, r *http.Request) {
	var dto RenameDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Rename(r.Context(), dto.PublicKey, dto.Name); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
