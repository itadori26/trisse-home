// Package controller contains the feed reader HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/rss/adapter/presenter"
	"trisse-home/internal/modules/rss/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the feed reader endpoints.
type Handlers struct{ uc usecase.Reader }

// NewHandlers returns feed reader handlers.
func NewHandlers(uc usecase.Reader) *Handlers { return &Handlers{uc: uc} }

// Feeds handles GET /api/rss/feeds.
func (h *Handlers) Feeds(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Feeds(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Feeds(res))
}

// Subscribe handles POST /api/rss/feeds.
func (h *Handlers) Subscribe(w http.ResponseWriter, r *http.Request) {
	var dto SubscribeDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Subscribe(r.Context(), dto.URL)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Feed(res))
}

// Unsubscribe handles DELETE /api/rss/feeds/{id}.
func (h *Handlers) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Unsubscribe(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Items handles GET /api/rss/items?feed=…&unread=true&limit=….
func (h *Handlers) Items(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit")) // invalid or missing means default
	res, err := h.uc.Items(r.Context(), usecase.ItemFilter{FeedID: q.Get("feed"), UnreadOnly: q.Get("unread") == "true", Limit: limit})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Items(res))
}

// SetRead handles PUT /api/rss/items/{id}/read.
func (h *Handlers) SetRead(w http.ResponseWriter, r *http.Request) {
	var dto ReadDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.MarkRead(r.Context(), r.PathValue("id"), dto.Read); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MarkAllRead handles POST /api/rss/read-all.
func (h *Handlers) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	var dto MarkAllReadDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.MarkAllRead(r.Context(), dto.FeedID); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Refresh handles POST /api/rss/refresh. Failures of single feeds are
// reported on the feeds themselves, so the request still succeeds.
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	_ = h.uc.RefreshAll(r.Context())
	h.Feeds(w, r)
}
