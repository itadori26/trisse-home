// Package controller contains the search HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/search/adapter/presenter"
	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the search endpoints.
type Handlers struct{ uc usecase.Searcher }

// NewHandlers returns search handlers.
func NewHandlers(uc usecase.Searcher) *Handlers { return &Handlers{uc: uc} }

// Search handles GET /api/search?q=… by redirecting to the resolved URL, so
// the search bar can be a plain form and browsers can use it as their engine.
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	target, err := h.uc.Resolve(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	// Not an open redirect: the route requires a session and the SameSite=Strict
	// cookie is never sent with cross-site navigations.
	http.Redirect(w, r, target, http.StatusFound) //nolint:gosec // see comment above
}

// Settings handles GET /api/search/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	h.respond(w, r, res, err)
}

// SetDefaultProvider handles PUT /api/search/settings.
func (h *Handlers) SetDefaultProvider(w http.ResponseWriter, r *http.Request) {
	var dto DefaultProviderDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SetDefaultProvider(r.Context(), dto.DefaultProvider)
	h.respond(w, r, res, err)
}

// AddShortcut handles POST /api/search/shortcuts.
func (h *Handlers) AddShortcut(w http.ResponseWriter, r *http.Request) {
	var dto presenter.ShortcutDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.AddShortcut(r.Context(), usecase.AddShortcutRequest{Keyword: dto.Keyword, Name: dto.Name, URLTemplate: dto.URLTemplate})
	h.respond(w, r, res, err)
}

// DeleteShortcut handles DELETE /api/search/shortcuts/{keyword}.
func (h *Handlers) DeleteShortcut(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.DeleteShortcut(r.Context(), r.PathValue("keyword"))
	h.respond(w, r, res, err)
}

// OpenSearch handles GET /api/search/opensearch.xml. Browsers fetch it
// without credentials, so it is public; it contains no personal data.
func (h *Handlers) OpenSearch(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	body, err := presenter.OpenSearch(scheme + "://" + r.Host)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/opensearchdescription+xml")
	_, _ = w.Write(body) //nolint:gosec // XML produced by encoding/xml, which escapes all values
}

func (h *Handlers) respond(w http.ResponseWriter, r *http.Request, res usecase.SettingsResponse, err error) {
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}
