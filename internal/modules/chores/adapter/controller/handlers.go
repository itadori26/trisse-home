// Package controller contains the chore plan HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/chores/adapter/presenter"
	"waylight/internal/modules/chores/domain"
	"waylight/internal/modules/chores/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the chore plan endpoints.
type Handlers struct{ uc usecase.Chores }

// NewHandlers returns chore plan handlers.
func NewHandlers(uc usecase.Chores) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/chores.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Chores(list))
}

func input(dto ChoreInputDTO) domain.Input {
	return domain.Input{Title: dto.Title, EveryDays: dto.EveryDays, Due: dto.Due, Assignees: dto.Assignees}
}

// Create handles POST /api/chores.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto ChoreInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Chore(v))
}

// Update handles PUT /api/chores/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto ChoreInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Chore(v))
}

// Delete handles DELETE /api/chores/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Done handles POST /api/chores/{id}/done.
func (h *Handlers) Done(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.Done(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Chore(v))
}
