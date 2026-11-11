package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/system/adapter/presenter"
	"trisse-home/internal/modules/system/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves GET /api/system.
type Handlers struct{ uc usecase.Monitor }

// NewHandlers returns system handlers.
func NewHandlers(uc usecase.Monitor) *Handlers { return &Handlers{uc: uc} }

// Snapshot handles GET /api/system.
func (h *Handlers) Snapshot(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Current(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Snapshot(res))
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
