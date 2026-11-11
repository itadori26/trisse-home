// Package controller contains the transit HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/transit/adapter/presenter"
	"trisse-home/internal/modules/transit/usecase"
	"trisse-home/internal/shared/httpapi"
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
