// Package controller contains the meal plan HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"waylight/internal/modules/meals/adapter/presenter"
	"waylight/internal/modules/meals/domain"
	"waylight/internal/modules/meals/usecase"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// defaultDays is one week, what the widget shows.
const defaultDays = 7

// Handlers serves the meal plan endpoints.
type Handlers struct{ uc usecase.Meals }

// NewHandlers returns meal plan handlers.
func NewHandlers(uc usecase.Meals) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/meals?from=YYYY-MM-DD&days=N.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	days := defaultDays
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("days", "Ask for 1 to 42 days."))
			return
		}
		days = n
	}
	list, err := h.uc.Range(r.Context(), r.URL.Query().Get("from"), days)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Entries(list))
}

func input(dto EntryInputDTO) domain.Input {
	return domain.Input{Date: dto.Date, Meal: dto.Meal, Dish: dto.Dish, Link: dto.Link, Ingredients: dto.Ingredients}
}

// Create handles POST /api/meals.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto EntryInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Entry(v))
}

// Update handles PUT /api/meals/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto EntryInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto), dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Entry(v))
}

// Delete handles DELETE /api/meals/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
