package controller

import (
	"net/http"

	"trisse-home/internal/modules/appearance/adapter/presenter"
	"trisse-home/internal/modules/appearance/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves /api/appearance.
type Handlers struct {
	get    usecase.GetAppearance
	update usecase.UpdateAppearance
}

// NewHandlers returns appearance handlers.
func NewHandlers(get usecase.GetAppearance, update usecase.UpdateAppearance) *Handlers {
	return &Handlers{get: get, update: update}
}

// Get handles GET /api/appearance.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.get.Execute(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Appearance(res))
}

// Update handles PUT /api/appearance.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto presenter.AppearanceDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.update.Execute(r.Context(), usecase.UpdateAppearanceRequest{
		Theme: dto.Theme, AccentColor: dto.AccentColor, BackgroundImage: dto.BackgroundImage, CustomCSS: dto.CustomCSS,
	})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Appearance(res))
}
