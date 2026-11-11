package controller

import (
	"net/http"
	"strconv"
	"time"

	"trisse-home/internal/modules/calendar/adapter/presenter"
	"trisse-home/internal/modules/calendar/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
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

// Events handles GET /api/calendar/events?days=… or ?from=…&to=… (RFC 3339).
func (h *Handlers) Events(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var res usecase.AgendaResponse
	var err error
	if q.Has("from") || q.Has("to") {
		from, errFrom := time.Parse(time.RFC3339, q.Get("from"))
		to, errTo := time.Parse(time.RFC3339, q.Get("to"))
		if errFrom != nil || errTo != nil {
			httpapi.WriteDomainError(w, r, kernel.NewValidationError("from", "from and to must be RFC 3339 timestamps."))
			return
		}
		res, err = h.uc.Range(r.Context(), from, to)
	} else {
		days, _ := strconv.Atoi(q.Get("days")) // invalid or missing means default
		res, err = h.uc.Upcoming(r.Context(), days)
	}
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Agenda(res))
}

// CreateEvent handles POST /api/calendar/events.
func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var dto NewEventDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	layout := time.RFC3339
	if dto.AllDay {
		layout = time.DateOnly
	}
	start, errStart := time.Parse(layout, dto.Start)
	end, errEnd := time.Parse(layout, dto.End)
	if errStart != nil || errEnd != nil {
		httpapi.WriteDomainError(w, r, kernel.NewValidationError("start", "Please enter a valid start and end."))
		return
	}
	res, err := h.uc.CreateEvent(r.Context(), usecase.NewEventInput{
		CalendarURL: dto.CalendarURL, Summary: dto.Summary, Location: dto.Location, AllDay: dto.AllDay, Start: start, End: end,
	})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Event(res))
}
