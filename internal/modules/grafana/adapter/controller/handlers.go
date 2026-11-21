// Package controller contains the Grafana HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"waylight/internal/modules/grafana/adapter/presenter"
	"waylight/internal/modules/grafana/domain"
	"waylight/internal/modules/grafana/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the Grafana endpoints.
type Handlers struct{ uc usecase.Grafana }

// NewHandlers returns Grafana handlers.
func NewHandlers(uc usecase.Grafana) *Handlers { return &Handlers{uc: uc} }

func respond[T any](w http.ResponseWriter, r *http.Request, v T, err error) {
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, v)
}

// Settings handles GET /api/grafana/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	respond(w, r, presenter.Settings(res), err)
}

// SaveSettings handles PUT /api/grafana/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput(dto))
	respond(w, r, presenter.Settings(res), err)
}

// connection is the named connection the widget chose (#610); empty means
// the module settings.
func connection(r *http.Request) string { return r.URL.Query().Get("connection") }

// Alerts handles GET /api/grafana/alerts[?connection=<id>].
func (h *Handlers) Alerts(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Alerts(r.Context(), connection(r))
	respond(w, r, presenter.Alerts(list, h.uc.URL(r.Context(), connection(r))), err)
}

// Dashboards handles GET /api/grafana/dashboards?q=….
func (h *Handlers) Dashboards(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Dashboards(r.Context(), connection(r), r.URL.Query().Get("q"))
	respond(w, r, presenter.Dashboards(list), err)
}

// Panels handles GET /api/grafana/dashboards/{uid}/panels.
func (h *Handlers) Panels(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Panels(r.Context(), connection(r), r.PathValue("uid"))
	respond(w, r, presenter.Panels(list), err)
}

// PanelData handles GET /api/grafana/data/{uid}/{panel}?range: the panel's
// values for a stat or sparkline when Grafana has no image renderer (#579).
func (h *Handlers) PanelData(w http.ResponseWriter, r *http.Request) {
	panel, _ := strconv.Atoi(r.PathValue("panel"))
	req, err := domain.NewDataRequest(r.PathValue("uid"), panel, r.URL.Query().Get("range"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	d, err := h.uc.PanelData(r.Context(), connection(r), req)
	respond(w, r, presenter.PanelData(d), err)
}

// Render handles GET /api/grafana/render/{uid}/{panel}?width&height&range&theme:
// the panel as PNG from Grafana's image renderer, so the browser never sees
// the token.
func (h *Handlers) Render(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	num := func(s string) int { n, _ := strconv.Atoi(s); return n }
	req, err := domain.NewRenderRequest(r.PathValue("uid"), num(r.PathValue("panel")), num(q.Get("width")), num(q.Get("height")),
		q.Get("range"), q.Get("theme") == "dark")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	img, err := h.uc.Render(r.Context(), connection(r), req)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", img.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=60")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(img.Data)
}
