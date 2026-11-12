package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"trisse-home/internal/modules/backup/usecase"
	"trisse-home/internal/shared/httpapi"
)

// maxFieldBytes limits the text fields of the restore form.
const maxFieldBytes = 1024

// RestoreHandlers serves POST /api/backups/restore (#46).
type RestoreHandlers struct {
	uc       usecase.Restorer
	maxBytes int64
}

// NewRestoreHandlers returns restore handlers; maxBytes limits the body.
func NewRestoreHandlers(uc usecase.Restorer, maxBytes int64) *RestoreHandlers {
	return &RestoreHandlers{uc: uc, maxBytes: maxBytes}
}

// Restore handles POST /api/backups/restore as multipart/form-data with the
// fields password, code, source (local|remote|upload), name and, last, the
// file. The file is streamed straight to the staging area; no temporary
// files are written elsewhere (the container's root file system is
// read-only).
func (h *RestoreHandlers) Restore(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		httpapi.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send the restore as multipart/form-data.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	mr, err := r.MultipartReader()
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
		return
	}
	req := usecase.RestoreRequest{ClientKey: httpapi.ClientIP(r.Context()).String()}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
			return
		}
		if part.FormName() == "file" {
			// The file must come last; it is consumed by the use case.
			req.Upload = part
			break
		}
		b, err := io.ReadAll(io.LimitReader(part, maxFieldBytes))
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "Invalid form data.")
			return
		}
		value := string(b)
		switch part.FormName() {
		case "password":
			req.Password = value
		case "code":
			req.Code = strings.TrimSpace(value)
		case "source":
			req.Source = usecase.RestoreSource(value)
		case "name":
			req.Name = value
		}
	}
	if err := h.uc.Restore(r.Context(), req); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "restarting"})
}
