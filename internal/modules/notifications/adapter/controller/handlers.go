// Package controller contains the notification HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"time"

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
		Policy: policyFrom(dto.Policy),
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

func policyFrom(dto PolicyDTO) domain.Policy {
	p := domain.Policy{Severities: map[string]domain.Severity{}}
	for k, v := range dto.Severities {
		p.Severities[k] = domain.Severity(v)
	}
	p.Quiet = domain.QuietHours{Enabled: dto.Quiet.Enabled, Start: dto.Quiet.Start, End: dto.Quiet.End}
	for _, d := range dto.Quiet.Days {
		if d >= 0 && d <= 6 {
			p.Quiet.Days = append(p.Quiet.Days, time.Weekday(d))
		}
	}
	return p
}

// Channels handles GET /api/notifications/channels.
func (h *Handlers) Channels(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Channels(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Channels(list))
}

func channelFrom(id string, dto ChannelDTO) usecase.ChannelInput {
	return usecase.ChannelInput{Channel: domain.Channel{ID: id, Kind: domain.ChannelKind(dto.Kind), Name: dto.Name, Enabled: dto.Enabled,
		MinSeverity: domain.Severity(dto.MinSeverity), Types: dto.Types, URL: dto.URL, Secret: dto.Secret, Template: dto.Template,
		SMTPHost: dto.SMTPHost, SMTPPort: dto.SMTPPort, Username: dto.Username, From: dto.From, To: dto.To, TLS: dto.TLS}}
}

// SaveChannel handles POST /api/notifications/channels and PUT …/{id}.
func (h *Handlers) SaveChannel(w http.ResponseWriter, r *http.Request) {
	var dto ChannelDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	c, err := h.uc.SaveChannel(r.Context(), channelFrom(r.PathValue("id"), dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Channel(c))
}

// DeleteChannel handles DELETE /api/notifications/channels/{id}.
func (h *Handlers) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteChannel(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// TestChannel handles POST /api/notifications/channels/{id}/test.
func (h *Handlers) TestChannel(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.TestChannel(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Maintenance handles GET /api/notifications/maintenance (active and upcoming).
func (h *Handlers) Maintenance(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Maintenance(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Maintenance(list))
}

// StartMaintenance handles POST /api/notifications/maintenance.
func (h *Handlers) StartMaintenance(w http.ResponseWriter, r *http.Request) {
	var dto MaintenanceDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	start := time.Now()
	if dto.Start != nil {
		start = *dto.Start
	}
	end := start.Add(time.Duration(dto.Minutes) * time.Minute)
	if dto.End != nil {
		end = *dto.End
	}
	win, err := h.uc.StartMaintenance(r.Context(), dto.Scope, start, end, dto.Note)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Maintenance([]domain.MaintenanceWindow{win})[0])
}

// EndMaintenance handles DELETE /api/notifications/maintenance/{id}.
func (h *Handlers) EndMaintenance(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.EndMaintenance(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
