// Package controller contains the holiday HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/holidays/adapter/presenter"
	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/modules/holidays/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves /api/holidays.
type Handlers struct{ uc usecase.Holidays }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Holidays) *Handlers { return &Handlers{uc: uc} }

// SettingsInputDTO is the body of PUT /api/holidays/settings.
type SettingsInputDTO struct {
	SchoolURL string `json:"schoolUrl"`
}

// Upcoming handles GET /api/holidays?state=NI&limit=8.
func (h *Handlers) Upcoming(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 8
	}
	o, err := h.uc.Upcoming(r.Context(), domain.ParseState(r.URL.Query().Get("state")), limit)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(o))
}

// Settings handles GET /api/holidays/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s))
}

// SaveSettings handles PUT /api/holidays/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.SaveSettings(r.Context(), dto.SchoolURL)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s))
}
