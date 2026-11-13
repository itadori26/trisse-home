package controller

import (
	"net/http"

	"trisse-home/internal/modules/calendar/adapter/presenter"
	"trisse-home/internal/modules/calendar/usecase"
	"trisse-home/internal/shared/httpapi"
)

// SubscriptionHandlers serve the ICS subscription endpoints (#98).
type SubscriptionHandlers struct{ uc usecase.Subscriptions }

// NewSubscriptionHandlers returns subscription handlers.
func NewSubscriptionHandlers(uc usecase.Subscriptions) *SubscriptionHandlers {
	return &SubscriptionHandlers{uc: uc}
}

// List handles GET /api/calendar/subscriptions.
func (h *SubscriptionHandlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.ListSubscriptions(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Subscriptions(res))
}

// Add handles POST /api/calendar/subscriptions.
func (h *SubscriptionHandlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto SubscriptionInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.AddSubscription(r.Context(), usecase.SubscriptionInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Subscription(res))
}

// Update handles PUT /api/calendar/subscriptions/{id}.
func (h *SubscriptionHandlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto SubscriptionInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.UpdateSubscription(r.Context(), r.PathValue("id"), usecase.SubscriptionInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Subscription(res))
}

// Delete handles DELETE /api/calendar/subscriptions/{id}.
func (h *SubscriptionHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteSubscription(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
