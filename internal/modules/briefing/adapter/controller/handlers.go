// Package controller contains the briefing HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/briefing/adapter/presenter"
	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/modules/briefing/usecase"
	"trisse-home/internal/shared/httpapi"
)

// UseCases bundles the briefing input ports.
type UseCases struct {
	Build    usecase.BuildBriefing
	Send     usecase.SendBriefing
	Settings usecase.ManageSettings
}

// Handlers serves the briefing endpoints.
type Handlers struct{ uc UseCases }

// NewHandlers returns briefing handlers.
func NewHandlers(uc UseCases) *Handlers { return &Handlers{uc: uc} }

// Get handles GET /api/briefing.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Build.Execute(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Briefing(res))
}

// GetSettings handles GET /api/briefing/settings.
func (h *Handlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings.Get(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/briefing/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in := usecase.SettingsInput{PushOn: dto.PushOn, PushAt: dto.PushAt, PushDays: dto.PushDays, SkipEmpty: dto.SkipEmpty, Lat: dto.Lat, Lon: dto.Lon}
	for _, e := range dto.Layout {
		in.Layout = append(in.Layout, domain.LayoutEntry{Kind: domain.SectionKind(e.Kind), Enabled: e.Enabled})
	}
	res, err := h.uc.Settings.Save(r.Context(), in)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Send handles POST /api/briefing/send: pushes the briefing now.
func (h *Handlers) Send(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Send.SendNow(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
