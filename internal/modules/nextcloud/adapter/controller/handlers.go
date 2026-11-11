// Package controller contains the Nextcloud HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/nextcloud/adapter/presenter"
	"trisse-home/internal/modules/nextcloud/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the Nextcloud endpoints.
type Handlers struct{ uc usecase.Nextcloud }

// NewHandlers returns Nextcloud handlers.
func NewHandlers(uc usecase.Nextcloud) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/nextcloud/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/nextcloud/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput{URL: dto.URL, Username: dto.Username, AppPassword: dto.AppPassword})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Status handles GET /api/nextcloud/status.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Status(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Status(res))
}
