// Package controller contains the summary HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"io"
	"net/http"
	"time"

	"trisse-home/internal/modules/summary/adapter/presenter"
	"trisse-home/internal/modules/summary/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves /api/summary.
type Handlers struct{ uc usecase.Summary }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Summary) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/summary/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(v))
}

// SaveSettings handles PUT /api/summary/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.SaveSettings(r.Context(), dto.URL, dto.Model, dto.MaxInputChars)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(v))
}

// Disable handles DELETE /api/summary/settings.
func (h *Handlers) Disable(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Disable(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Models handles POST /api/summary/models.
func (h *Handlers) Models(w http.ResponseWriter, r *http.Request) {
	var dto ModelsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	models, err := h.uc.Models(r.Context(), dto.URL)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ModelsDTO{Models: models})
}

// Summarize handles POST /api/summary.
func (h *Handlers) Summarize(w http.ResponseWriter, r *http.Request) {
	var dto SummarizeDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Summarize(r.Context(), usecase.Request(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Result(res))
}

// askTimeout bounds one streamed answer; the server's write timeout would
// otherwise cut a slow model off.
const askTimeout = 3 * time.Minute

// Ask handles POST /api/summary/ask: the answer is streamed as plain text.
// Errors before the first chunk are ordinary JSON errors; a closed
// connection cancels the model request.
func (h *Handlers) Ask(w http.ResponseWriter, r *http.Request) {
	var dto AskDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Now().Add(askTimeout))
	started := false
	err := h.uc.Ask(r.Context(), dto.Question, dto.Context, dto.Language, func(chunk string) error {
		if !started {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusOK)
			started = true
		}
		if _, err := io.WriteString(w, chunk); err != nil {
			return err
		}
		return rc.Flush()
	})
	if err != nil && !started {
		httpapi.WriteDomainError(w, r, err)
	}
	if !started && err == nil {
		w.WriteHeader(http.StatusNoContent)
	}
}
