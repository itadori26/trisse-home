// Package controller contains the transit HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"waylight/internal/modules/transit/adapter/presenter"
	"waylight/internal/modules/transit/domain"
	"waylight/internal/modules/transit/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the transit endpoints.
type Handlers struct{ uc usecase.Transit }

// NewHandlers returns transit handlers.
func NewHandlers(uc usecase.Transit) *Handlers { return &Handlers{uc: uc} }

// Stops handles GET /api/transit/stops?q=….
func (h *Handlers) Stops(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.SearchStops(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Stops(res))
}

// StopsNear handles GET /api/transit/stops/near?lat=…&lon=….
func (h *Handlers) StopsNear(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat, errLat := strconv.ParseFloat(q.Get("lat"), 64)
	lon, errLon := strconv.ParseFloat(q.Get("lon"), 64)
	if errLat != nil || errLon != nil {
		httpapi.WriteDomainError(w, r, domain.ErrInvalidLocation)
		return
	}
	res, err := h.uc.StopsNear(r.Context(), lat, lon)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Stops(res))
}

// Departures handles GET /api/transit/departures?stop=…&n=….
func (h *Handlers) Departures(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n")) // invalid → clamped by the domain
	res, err := h.uc.Departures(r.Context(), r.URL.Query().Get("stop"), n)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Departures(res))
}
