// Package controller contains the shopping list HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/shopping/adapter/presenter"
	"trisse-home/internal/modules/shopping/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the shopping list endpoints.
type Handlers struct{ uc usecase.Shopping }

// NewHandlers returns shopping list handlers.
func NewHandlers(uc usecase.Shopping) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/shopping/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/shopping/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), dto.ListURL, dto.Username, dto.Password)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// List handles GET /api/shopping/items.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Items(items))
}

// Add handles POST /api/shopping/items.
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto NewItemDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	it, err := h.uc.Add(r.Context(), dto.Summary)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Item(it))
}

// SetDone handles PUT /api/shopping/items/{id}.
func (h *Handlers) SetDone(w http.ResponseWriter, r *http.Request) {
	var dto ItemStateDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.SetDone(r.Context(), r.PathValue("id"), dto.Done); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/shopping/items/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClearCompleted handles POST /api/shopping/clear-completed.
func (h *Handlers) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.ClearCompleted(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ClearedDTO{Deleted: n})
}
