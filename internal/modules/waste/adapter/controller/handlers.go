// Package controller contains the waste HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/waste/adapter/presenter"
	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/modules/waste/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the waste endpoints.
type Handlers struct{ uc usecase.Waste }

// NewHandlers returns waste handlers.
func NewHandlers(uc usecase.Waste) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/waste/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	s, configured, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s, configured))
}

// SaveSettings handles PUT /api/waste/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	rules := make([]domain.Rule, 0, len(dto.Rules))
	for _, x := range dto.Rules {
		rules = append(rules, domain.Rule(x))
	}
	s, err := h.uc.SaveSettings(r.Context(), dto.URL, rules)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s, true))
}

// Upcoming handles GET /api/waste/upcoming.
func (h *Handlers) Upcoming(w http.ResponseWriter, r *http.Request) {
	o, err := h.uc.Upcoming(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(o))
}
