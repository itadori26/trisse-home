package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/system/adapter/presenter"
	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/modules/system/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves GET /api/system and the disk forecast settings.
type Handlers struct {
	uc        usecase.Monitor
	forecasts usecase.DiskForecasts
}

// NewHandlers returns system handlers.
func NewHandlers(uc usecase.Monitor, forecasts usecase.DiskForecasts) *Handlers {
	return &Handlers{uc: uc, forecasts: forecasts}
}

// Snapshot handles GET /api/system.
func (h *Handlers) Snapshot(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Current(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	// The forecast is an extra; the snapshot is shown without it on errors.
	forecasts, _ := h.forecasts.Forecasts(r.Context())
	httpapi.WriteJSON(w, http.StatusOK, presenter.Snapshot(res, forecasts))
}

// ForecastSettings handles GET /api/system/forecast-settings.
func (h *Handlers) ForecastSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.forecasts.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ForecastSettingsDTO{WindowDays: s.WindowDays, AlertDays: s.AlertDays})
}

// SaveForecastSettings handles PUT /api/system/forecast-settings.
func (h *Handlers) SaveForecastSettings(w http.ResponseWriter, r *http.Request) {
	var dto ForecastSettingsDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.forecasts.SaveSettings(r.Context(), domain.ForecastSettings{WindowDays: dto.WindowDays, AlertDays: dto.AlertDays})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ForecastSettingsDTO{WindowDays: s.WindowDays, AlertDays: s.AlertDays})
}

// History handles GET /api/system/history?hours=….
func (h *Handlers) History(w http.ResponseWriter, r *http.Request) {
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours")) // invalid or missing means the full range
	res, err := h.uc.History(r.Context(), hours)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.History(res))
}
