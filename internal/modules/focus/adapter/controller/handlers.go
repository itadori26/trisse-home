// Package controller contains the focus timer HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/focus/adapter/presenter"
	"trisse-home/internal/modules/focus/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the focus timer endpoints.
type Handlers struct{ uc usecase.Focus }

// NewHandlers returns focus timer handlers.
func NewHandlers(uc usecase.Focus) *Handlers { return &Handlers{uc: uc} }

// Record handles POST /api/focus/sessions and answers with the updated
// statistics.
func (h *Handlers) Record(w http.ResponseWriter, r *http.Request) {
	var dto SessionInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if _, err := h.uc.RecordSession(r.Context(), dto.Minutes, dto.EndedAt); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.write(w, r, http.StatusCreated)
}

// Stats handles GET /api/focus/stats.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) { h.write(w, r, http.StatusOK) }

func (h *Handlers) write(w http.ResponseWriter, r *http.Request, status int) {
	days, err := h.uc.WeeklyStats(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, status, presenter.Week(days))
}
