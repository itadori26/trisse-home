// Package controller contains the gift idea HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"waylight/internal/modules/gifts/adapter/presenter"
	"waylight/internal/modules/gifts/domain"
	"waylight/internal/modules/gifts/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the gift idea endpoints.
type Handlers struct{ uc usecase.Gifts }

// NewHandlers returns gift idea handlers.
func NewHandlers(uc usecase.Gifts) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/gifts.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Ideas(list))
}

func input(dto IdeaInputDTO) domain.Input {
	return domain.Input{Person: dto.Person, Title: dto.Title, Link: dto.Link, Note: dto.Note, Bought: dto.Bought}
}

// Create handles POST /api/gifts.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto IdeaInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	i, err := h.uc.Create(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Idea(i))
}

// Update handles PUT /api/gifts/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto IdeaInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	i, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Idea(i))
}

// Delete handles DELETE /api/gifts/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
