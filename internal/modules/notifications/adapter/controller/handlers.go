// Package controller contains the notification HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/notifications/adapter/presenter"
	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/modules/notifications/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the notification endpoints.
type Handlers struct{ uc usecase.Notifier }

// NewHandlers returns notification handlers.
func NewHandlers(uc usecase.Notifier) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/notifications/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/notifications/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput{
		Enabled: dto.Enabled, ServerURL: dto.ServerURL, Topic: dto.Topic, Token: dto.Token, Rules: domain.Rules(dto.Rules),
	})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SendTest handles POST /api/notifications/test.
func (h *Handlers) SendTest(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.SendTest(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
