// Package controller contains the countdown HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/countdown/adapter/presenter"
	"trisse-home/internal/modules/countdown/domain"
	"trisse-home/internal/modules/countdown/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// DefaultGraceDays is how long past countdowns stay visible when the
// widget does not say otherwise.
const DefaultGraceDays = 1

// Handlers serves the countdown endpoints.
type Handlers struct{ uc usecase.Countdowns }

// NewHandlers returns countdown handlers.
func NewHandlers(uc usecase.Countdowns) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/countdowns?graceDays=N.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	grace := DefaultGraceDays
	if v := r.URL.Query().Get("graceDays"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("graceDays", "graceDays must be a number."))
			return
		}
		grace = n
	}
	list, err := h.uc.List(r.Context(), grace)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Countdowns(list))
}

func input(dto CountdownInputDTO) domain.Input {
	return domain.Input{Title: dto.Title, Emoji: dto.Emoji, Date: dto.Date, Time: dto.Time}
}

// Create handles POST /api/countdowns.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto CountdownInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	c, err := h.uc.Create(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Countdown(c))
}

// Update handles PUT /api/countdowns/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto CountdownInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	c, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Countdown(c))
}

// Delete handles DELETE /api/countdowns/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
