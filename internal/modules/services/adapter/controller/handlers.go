package controller

import (
	"net/http"

	"trisse-home/internal/modules/services/adapter/presenter"
	"trisse-home/internal/modules/services/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the services endpoints.
type Handlers struct{ uc usecase.Services }

// NewHandlers returns services handlers.
func NewHandlers(uc usecase.Services) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/services.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Services(res))
}

// Create handles POST /api/services.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto ServiceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Create(r.Context(), dto.toInput())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Service(res))
}

// Update handles PUT /api/services/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto ServiceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Update(r.Context(), r.PathValue("id"), dto.toInput())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Service(res))
}

// Delete handles DELETE /api/services/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reorder handles PUT /api/services/order.
func (h *Handlers) Reorder(w http.ResponseWriter, r *http.Request) {
	var dto OrderDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Reorder(r.Context(), dto.IDs); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Check handles POST /api/services/check (manual refresh).
func (h *Handlers) Check(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.CheckAll(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.List(w, r)
}
