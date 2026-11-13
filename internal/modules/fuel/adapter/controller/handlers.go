// Package controller contains the fuel price HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/fuel/adapter/presenter"
	"trisse-home/internal/modules/fuel/domain"
	"trisse-home/internal/modules/fuel/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the fuel price endpoints.
type Handlers struct{ uc usecase.Fuel }

// NewHandlers returns fuel handlers.
func NewHandlers(uc usecase.Fuel) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/fuel/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	h.writeSettings(w, r)(h.uc.Settings(r.Context()))
}

// SaveSettings handles PUT /api/fuel/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.writeSettings(w, r)(h.uc.SaveSettings(r.Context(), usecase.SettingsInput(dto)))
}

// Favourites handles PUT /api/fuel/favourites.
func (h *Handlers) Favourites(w http.ResponseWriter, r *http.Request) {
	var dto []FavouriteDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	list := make([]domain.Favourite, 0, len(dto))
	for _, f := range dto {
		list = append(list, domain.Favourite(f))
	}
	h.writeSettings(w, r)(h.uc.SetFavourites(r.Context(), list))
}

// Overview handles GET /api/fuel/overview.
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Overview(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(res))
}

func (h *Handlers) writeSettings(w http.ResponseWriter, r *http.Request) func(usecase.SettingsView, error) {
	return func(v usecase.SettingsView, err error) {
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(v))
	}
}
