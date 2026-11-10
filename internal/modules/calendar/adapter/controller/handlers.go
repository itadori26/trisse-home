package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/calendar/adapter/presenter"
	"trisse-home/internal/modules/calendar/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the calendar endpoints.
type Handlers struct{ uc usecase.Calendar }

// NewHandlers returns calendar handlers.
func NewHandlers(uc usecase.Calendar) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/calendar/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/calendar/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput{CalendarURLs: dto.CalendarURLs, Username: dto.Username, Password: dto.Password})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Events handles GET /api/calendar/events?days=….
func (h *Handlers) Events(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days")) // invalid or missing means default
	res, err := h.uc.Upcoming(r.Context(), days)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Agenda(res))
}
