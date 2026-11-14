// Package controller contains the waste HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

	"trisse-home/internal/modules/waste/adapter/presenter"
	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/modules/waste/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the waste endpoints.
type Handlers struct{ uc usecase.Waste }

// NewHandlers returns waste handlers.
func NewHandlers(uc usecase.Waste) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/waste/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	s, configured, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s, configured))
}

// SaveSettings handles PUT /api/waste/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	rules := make([]domain.Rule, 0, len(dto.Rules))
	for _, x := range dto.Rules {
		rules = append(rules, domain.Rule(x))
	}
	s, err := h.uc.SaveSettings(r.Context(), dto.URL, rules)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s, true))
}

// Upcoming handles GET /api/waste/upcoming.
func (h *Handlers) Upcoming(w http.ResponseWriter, r *http.Request) {
	o, err := h.uc.Upcoming(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(o))
}

// uploadOverhead is room for the multipart headers around the file.
const uploadOverhead = 16 << 10

// UploadCalendar handles POST /api/waste/calendar (#153): multipart/form-data
// with one field "file", an .ics file or text/calendar of at most 1 MB.
func (h *Handlers) UploadCalendar(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the calendar as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxCalendarBytes+uploadOverhead)
	mr, err := r.MultipartReader()
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
		return
	}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeUploadError(w, err)
			return
		}
		if part.FormName() != "file" {
			continue
		}
		if !isICS(part.FileName(), part.Header.Get("Content-Type")) {
			httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Upload an .ics file (text/calendar).")
			return
		}
		content, err := io.ReadAll(io.LimitReader(part, domain.MaxCalendarBytes+1))
		if err != nil {
			writeUploadError(w, err)
			return
		}
		if len(content) > domain.MaxCalendarBytes {
			httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The file is larger than 1 MB.")
			return
		}
		s, err := h.uc.UploadCalendar(r.Context(), part.FileName(), content)
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s, true))
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The form has no file.")
}

// RemoveCalendar handles DELETE /api/waste/calendar.
func (h *Handlers) RemoveCalendar(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.RemoveCalendar(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s, s.Source.Configured()))
}

// isICS accepts files named *.ics or sent as text/calendar; browsers send
// .ics files with varying types (text/calendar, application/octet-stream).
func isICS(name, contentType string) bool {
	if strings.EqualFold(path.Ext(name), ".ics") {
		return true
	}
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/calendar"
}

func writeUploadError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The file is larger than 1 MB.")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
}
