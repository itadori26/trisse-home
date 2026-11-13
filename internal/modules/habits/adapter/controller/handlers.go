// Package controller contains the habit HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/habits/adapter/presenter"
	"trisse-home/internal/modules/habits/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the habit endpoints.
type Handlers struct{ uc usecase.Habits }

// NewHandlers returns habit handlers.
func NewHandlers(uc usecase.Habits) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/habits.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Habits(list))
}

func input(dto HabitInputDTO) usecase.Input {
	return usecase.Input{Name: dto.Name, Emoji: dto.Emoji, Days: dto.Days}
}

// Create handles POST /api/habits.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto HabitInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Habit(v))
}

// Update handles PUT /api/habits/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto HabitInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Habit(v))
}

// Delete handles DELETE /api/habits/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Toggle handles POST /api/habits/{id}/toggle.
func (h *Handlers) Toggle(w http.ResponseWriter, r *http.Request) {
	var dto ToggleDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Toggle(r.Context(), r.PathValue("id"), dto.Date)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Habit(v))
}
