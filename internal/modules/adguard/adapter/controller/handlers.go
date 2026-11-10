package controller

import (
	"net/http"

	"trisse-home/internal/modules/adguard/adapter/presenter"
	"trisse-home/internal/modules/adguard/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the AdGuard endpoints.
type Handlers struct{ uc usecase.AdGuard }

// NewHandlers returns AdGuard handlers.
func NewHandlers(uc usecase.AdGuard) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/adguard/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/adguard/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput{URL: dto.URL, Username: dto.Username, Password: dto.Password})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Stats handles GET /api/adguard/stats.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Stats(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Stats(res))
}
