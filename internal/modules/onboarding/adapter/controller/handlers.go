// Package controller contains the onboarding HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/onboarding/adapter/presenter"
	"trisse-home/internal/modules/onboarding/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves /api/onboarding.
type Handlers struct{ uc usecase.Onboarding }

// NewHandlers returns onboarding handlers.
func NewHandlers(uc usecase.Onboarding) *Handlers { return &Handlers{uc: uc} }

// Status handles GET /api/onboarding.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Status(r.Context())
	h.respond(w, r, res, err)
}

// Update handles PUT /api/onboarding.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto UpdateDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	var (
		res usecase.StatusResponse
		err error
	)
	if dto.Completed {
		res, err = h.uc.Complete(r.Context())
	} else {
		res, err = h.uc.Restart(r.Context())
	}
	h.respond(w, r, res, err)
}

func (h *Handlers) respond(w http.ResponseWriter, r *http.Request, res usecase.StatusResponse, err error) {
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Status(res))
}
