// Package controller contains the parcel tracking HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/parcels/adapter/presenter"
	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the parcel endpoints.
type Handlers struct{ uc usecase.Parcels }

// NewHandlers returns parcel handlers.
func NewHandlers(uc usecase.Parcels) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/parcels/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/parcels/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), dto.DHLAPIKey, dto.HideDeliveredDays)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// List handles GET /api/parcels.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	ps, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Parcels(ps))
}

// Add handles POST /api/parcels.
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto NewParcelDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	p, err := h.uc.Add(r.Context(), dto.TrackingNumber, domain.Carrier(dto.Carrier), dto.Label)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Parcel(p))
}

// Remove handles DELETE /api/parcels/{id}.
func (h *Handlers) Remove(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Remove(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MailSettings handles GET /api/parcels/mail (#166).
func (h *Handlers) MailSettings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.MailSettings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.MailSettings(res))
}

// SaveMailSettings handles PUT /api/parcels/mail; an enabled mailbox is
// tested before it is saved.
func (h *Handlers) SaveMailSettings(w http.ResponseWriter, r *http.Request) {
	var dto MailSettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveMailSettings(r.Context(), usecase.MailSettingsInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.MailSettings(res))
}
