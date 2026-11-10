package controller

import (
	"net/http"

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
