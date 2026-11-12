package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/speedtest/adapter/presenter"
	"trisse-home/internal/modules/speedtest/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the speedtest endpoints.
type Handlers struct{ uc usecase.Speedtest }

// NewHandlers returns speedtest handlers.
func NewHandlers(uc usecase.Speedtest) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/speedtest/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/speedtest/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput{URL: dto.URL, Token: dto.Token, MinDownMbps: dto.MinDownMbps, MinUpMbps: dto.MinUpMbps})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Results handles GET /api/speedtest/results?hours=24|168.
func (h *Handlers) Results(w http.ResponseWriter, r *http.Request) {
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours")) // invalid or missing means 7 days
	res, err := h.uc.Results(r.Context(), hours)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Results(res))
}
