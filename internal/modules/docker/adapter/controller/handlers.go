// Package controller contains the docker HTTP handler.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/docker/adapter/presenter"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves GET /api/docker/containers.
type Handlers struct{ uc usecase.Containers }

// NewHandlers returns docker handlers.
func NewHandlers(uc usecase.Containers) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/docker/containers.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(res))
}
