// Package controller contains the FRITZ!Box HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/fritzbox/adapter/presenter"
	"trisse-home/internal/modules/fritzbox/domain"
	"trisse-home/internal/modules/fritzbox/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// Handlers serves the FRITZ!Box endpoints.
type Handlers struct{ uc usecase.FritzBox }

// NewHandlers returns FRITZ!Box handlers.
func NewHandlers(uc usecase.FritzBox) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/fritzbox/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/fritzbox/settings.
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

// Status handles GET /api/fritzbox/status.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Status(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Status(res))
}

// Calls handles GET /api/fritzbox/calls?filter=missed&limit=10 (#168).
func (h *Handlers) Calls(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 0
	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			httpapi.WriteDomainError(w, r, kernel.NewValidationErrorf("limit", "Show between {min} and {max} calls.", "min", 5, "max", domain.MaxCalls))
			return
		}
		limit = n
	}
	f, err := domain.NewCallFilter(q.Get("filter"), limit)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Calls(r.Context(), f)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Calls(res))
}
