// Package controller contains the weather HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/weather/adapter/presenter"
	"trisse-home/internal/modules/weather/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// Handlers serves the weather endpoints.
type Handlers struct{ uc usecase.Weather }

// NewHandlers returns weather handlers.
func NewHandlers(uc usecase.Weather) *Handlers { return &Handlers{uc: uc} }

// Forecast handles GET /api/weather?lat=…&lon=….
func (h *Handlers) Forecast(w http.ResponseWriter, r *http.Request) {
	lat, errLat := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, errLon := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if errLat != nil || errLon != nil {
		httpapi.WriteDomainError(w, r, kernel.NewValidationError("lat", "Latitude and longitude are required."))
		return
	}
	res, err := h.uc.GetForecast(r.Context(), usecase.ForecastRequest{Lat: lat, Lon: lon})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Forecast(res))
}

// Places handles GET /api/weather/places?q=….
func (h *Handlers) Places(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.SearchPlaces(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Places(res))
}
