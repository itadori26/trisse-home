// Package controller contains the drop zone HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"trisse-home/internal/modules/drop/adapter/presenter"
	"trisse-home/internal/modules/drop/domain"
	"trisse-home/internal/modules/drop/usecase"
	"trisse-home/internal/shared/httpapi"
)

// TextDTO is the body of POST /api/drop/text; expiresInHours is 1, 24 or 168 (0 = 24).
type TextDTO struct {
	Text           string `json:"text"`
	ExpiresInHours int    `json:"expiresInHours"`
}

// Handlers serves /api/drop.
type Handlers struct{ uc usecase.Drop }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Drop) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/drop.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Items(items))
}

// AddText handles POST /api/drop/text.
func (h *Handlers) AddText(w http.ResponseWriter, r *http.Request) {
	var dto TextDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	i, err := h.uc.AddText(r.Context(), dto.Text, time.Duration(dto.ExpiresInHours)*time.Hour)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Item(i))
}

// uploadOverhead is room for the multipart headers around the file.
const uploadOverhead = 64 << 10

// AddFile handles POST /api/drop/files: multipart with "expiresInHours"
// (optional, before the file) and "file".
func (h *Handlers) AddFile(w http.ResponseWriter, r *http.Request) {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the file as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxFileBytes+uploadOverhead)
	mr, err := r.MultipartReader()
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
		return
	}
	hours := 0
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeUploadError(w, err)
			return
		}
		switch part.FormName() {
		case "expiresInHours":
			b, _ := io.ReadAll(io.LimitReader(part, 8))
			hours, _ = strconv.Atoi(string(b))
		case "file":
			data, err := io.ReadAll(io.LimitReader(part, domain.MaxFileBytes+1))
			if err != nil {
				writeUploadError(w, err)
				return
			}
			ct, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
			i, err := h.uc.AddFile(r.Context(), part.FileName(), ct, data, time.Duration(hours)*time.Hour)
			if err != nil {
				httpapi.WriteDomainError(w, r, err)
				return
			}
			httpapi.WriteJSON(w, http.StatusCreated, presenter.Item(i))
			return
		}
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The form has no file.")
}

func writeUploadError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The file must be at most 25 MB.")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
}

// Delete handles DELETE /api/drop/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// File handles GET /api/drop/{id}/file: always a download, passive types
// keep their content type, everything else is application/octet-stream.
func (h *Handlers) File(w http.ResponseWriter, r *http.Request) {
	i, data, err := h.uc.File(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", domain.SafeInlineType(i.ContentType))
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": i.FileName})+
		"; filename*=UTF-8''"+url.PathEscape(i.FileName))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(data)
}
