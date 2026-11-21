// Package controller contains the *arr HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/arr/adapter/presenter"
	"waylight/internal/modules/arr/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the *arr endpoints.
type Handlers struct{ uc usecase.Arr }

// NewHandlers returns *arr handlers.
func NewHandlers(uc usecase.Arr) *Handlers { return &Handlers{uc: uc} }

// Instances handles GET /api/arr/instances.
func (h *Handlers) Instances(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Instances(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Instances(res))
}

// Add handles POST /api/arr/instances.
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto InstanceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.AddInstance(r.Context(), usecase.InstanceInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Instance(res))
}

// Update handles PUT /api/arr/instances/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto InstanceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.UpdateInstance(r.Context(), r.PathValue("id"), usecase.InstanceInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Instance(res))
}

// Delete handles DELETE /api/arr/instances/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteInstance(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Overview handles GET /api/arr/overview[?connection=<id>&connection=<id>…]:
// a widget may choose a Sonarr, a Radarr and a Prowlarr connection (#610).
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Overview(r.Context(), r.URL.Query()["connection"])
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(res))
}
