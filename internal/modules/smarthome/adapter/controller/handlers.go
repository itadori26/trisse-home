// Package controller contains the smart home HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strings"

	"trisse-home/internal/modules/smarthome/adapter/presenter"
	"trisse-home/internal/modules/smarthome/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the smart home endpoints.
type Handlers struct{ uc usecase.SmartHome }

// NewHandlers returns smart home handlers.
func NewHandlers(uc usecase.SmartHome) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/smarthome/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/smarthome/settings.
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

// Entities handles GET /api/smarthome/entities.
func (h *Handlers) Entities(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Entities(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Entities(res))
}

// States handles GET /api/smarthome/states?ids=a.b,c.d.
func (h *Handlers) States(w http.ResponseWriter, r *http.Request) {
	var ids []string
	for id := range strings.SplitSeq(r.URL.Query().Get("ids"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	res, err := h.uc.States(r.Context(), ids)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Entities(res))
}

// Act handles POST /api/smarthome/entities/{id}/action.
func (h *Handlers) Act(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Act(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
