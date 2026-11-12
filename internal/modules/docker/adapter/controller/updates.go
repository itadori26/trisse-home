package controller

import (
	"net/http"

	"trisse-home/internal/modules/docker/adapter/presenter"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/shared/httpapi"
)

// UpdateHandlers serves the image update hints (#107).
type UpdateHandlers struct{ uc usecase.ImageUpdates }

// NewUpdateHandlers returns update handlers.
func NewUpdateHandlers(uc usecase.ImageUpdates) *UpdateHandlers { return &UpdateHandlers{uc: uc} }

// List handles GET /api/docker/updates.
func (h *UpdateHandlers) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ImageUpdates(list))
}

// Settings handles GET /api/docker/updates/settings.
func (h *UpdateHandlers) Settings(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.UpdateSettings(v))
}

// SaveSettings handles PUT /api/docker/updates/settings.
func (h *UpdateHandlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto UpdateSettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in := usecase.UpdateSettingsInput{Enabled: dto.Enabled, SemverTags: dto.SemverTags}
	for _, c := range dto.Registries {
		in.Credentials = append(in.Credentials, usecase.RegistryCredential(c))
	}
	v, err := h.uc.SaveSettings(r.Context(), in)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.UpdateSettings(v))
}
