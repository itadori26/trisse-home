// Package controller contains the game deals HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/gamedeals/adapter/presenter"
	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/modules/gamedeals/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the game deals endpoints.
type Handlers struct{ uc usecase.GameDeals }

// NewHandlers returns handlers.
func NewHandlers(uc usecase.GameDeals) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/gamedeals/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s))
}

// SaveSettings handles PUT /api/gamedeals/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.SaveSettings(r.Context(), domain.Settings(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s))
}

// Overview handles GET /api/gamedeals/overview.
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	o, err := h.uc.Overview(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(o))
}
