package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/weather/adapter/presenter"
	"trisse-home/internal/modules/weather/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// AlertHandlers serves warnings and the rain radar.
type AlertHandlers struct{ uc usecase.Alerts }

// NewAlertHandlers returns alert handlers.
func NewAlertHandlers(uc usecase.Alerts) *AlertHandlers { return &AlertHandlers{uc: uc} }

func coordinates(r *http.Request) (float64, float64, error) {
	lat, errLat := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, errLon := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if errLat != nil || errLon != nil {
		return 0, 0, kernel.NewValidationError("lat", "Latitude and longitude are required.")
	}
	return lat, lon, nil
}

// Warnings handles GET /api/weather/warnings?lat=…&lon=….
func (h *AlertHandlers) Warnings(w http.ResponseWriter, r *http.Request) {
	lat, lon, err := coordinates(r)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	ws, err := h.uc.Warnings(r.Context(), lat, lon)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Warnings(ws))
}

// Radar handles GET /api/weather/radar?lat=…&lon=… and returns a PNG.
func (h *AlertHandlers) Radar(w http.ResponseWriter, r *http.Request) {
	lat, lon, err := coordinates(r)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	img, err := h.uc.Radar(r.Context(), lat, lon)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("Last-Modified", img.FetchedAt.UTC().Format(http.TimeFormat))
	_, _ = w.Write(img.PNG)
}
