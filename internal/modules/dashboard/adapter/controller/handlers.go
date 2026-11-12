package controller

import (
	"bytes"
	"encoding/json"
	"net/http"

	"trisse-home/internal/modules/dashboard/adapter/presenter"
	"trisse-home/internal/modules/dashboard/domain"
	"trisse-home/internal/modules/dashboard/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// Handlers serves the dashboard endpoints.
type Handlers struct {
	get       usecase.GetDashboard
	tabs      usecase.ManageTabs
	widgets   usecase.ManageWidgets
	templates usecase.ManageTemplates
}

// NewHandlers returns dashboard handlers.
func NewHandlers(get usecase.GetDashboard, tabs usecase.ManageTabs, widgets usecase.ManageWidgets, templates usecase.ManageTemplates) *Handlers {
	return &Handlers{get: get, tabs: tabs, widgets: widgets, templates: templates}
}

// Get handles GET /api/dashboard.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	// The client address lets network rules pick the default tab (#71).
	res, err := h.get.ExecuteFor(r.Context(), httpapi.ClientIP(r.Context()))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Dashboard(res))
}

// CreateTab handles POST /api/dashboard/tabs.
func (h *Handlers) CreateTab(w http.ResponseWriter, r *http.Request) {
	var dto TabNameDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.tabs.Create(r.Context(), usecase.CreateTabRequest{Name: dto.Name})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Tab(res))
}

// RenameTab handles PATCH /api/dashboard/tabs/{id}.
func (h *Handlers) RenameTab(w http.ResponseWriter, r *http.Request) {
	var dto TabNameDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.tabs.Rename(r.Context(), usecase.RenameTabRequest{TabID: r.PathValue("id"), Name: dto.Name})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Tab(res))
}

// DeleteTab handles DELETE /api/dashboard/tabs/{id}.
func (h *Handlers) DeleteTab(w http.ResponseWriter, r *http.Request) {
	if err := h.tabs.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReorderTabs handles PUT /api/dashboard/tabs/order.
func (h *Handlers) ReorderTabs(w http.ResponseWriter, r *http.Request) {
	var dto TabOrderDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.tabs.Reorder(r.Context(), usecase.ReorderTabsRequest{TabIDs: dto.TabIDs}); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddWidget handles POST /api/dashboard/tabs/{id}/widgets.
func (h *Handlers) AddWidget(w http.ResponseWriter, r *http.Request) {
	var dto AddWidgetDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	config, err := normaliseConfig(dto.Config)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.widgets.Add(r.Context(), usecase.AddWidgetRequest{
		TabID: r.PathValue("id"), Kind: dto.Kind, X: dto.X, Y: dto.Y, W: dto.W, H: dto.H, Config: config,
	})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Widget(res))
}

// ConfigureWidget handles PATCH /api/dashboard/widgets/{id}.
func (h *Handlers) ConfigureWidget(w http.ResponseWriter, r *http.Request) {
	var dto WidgetConfigDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	config, err := normaliseConfig(dto.Config)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.widgets.Configure(r.Context(), usecase.ConfigureWidgetRequest{WidgetID: r.PathValue("id"), Config: config})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Widget(res))
}

// SaveLayout handles PUT /api/dashboard/tabs/{id}/layout.
func (h *Handlers) SaveLayout(w http.ResponseWriter, r *http.Request) {
	var dto LayoutDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	req := usecase.SaveLayoutRequest{TabID: r.PathValue("id"), Placements: make([]usecase.WidgetPlacement, 0, len(dto.Items))}
	for _, it := range dto.Items {
		req.Placements = append(req.Placements, usecase.WidgetPlacement{WidgetID: it.ID, X: it.X, Y: it.Y, W: it.W, H: it.H})
	}
	if err := h.widgets.SaveLayout(r.Context(), req); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteWidget handles DELETE /api/dashboard/widgets/{id}.
func (h *Handlers) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	if err := h.widgets.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// normaliseConfig ensures the configuration is a JSON object and compacts it.
func normaliseConfig(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "{}", nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", kernel.NewValidationError("config", "The widget configuration must be a JSON object.")
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return "", kernel.NewValidationError("config", "The widget configuration must be a JSON object.")
	}
	return buf.String(), nil
}

// SetBackground handles PUT /api/dashboard/tabs/{id}/background.
func (h *Handlers) SetBackground(w http.ResponseWriter, r *http.Request) {
	var dto TabBackgroundDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.tabs.SetBackground(r.Context(), r.PathValue("id"), dto.Background)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Tab(res))
}

// SetRules handles PUT /api/dashboard/tabs/{id}/rules.
func (h *Handlers) SetRules(w http.ResponseWriter, r *http.Request) {
	var dto TabRulesDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in := make([]domain.RuleInput, 0, len(dto.Rules))
	for _, x := range dto.Rules {
		in = append(in, domain.RuleInput{Days: x.Days, From: x.From, To: x.To, Networks: x.Networks})
	}
	res, err := h.tabs.SetRules(r.Context(), r.PathValue("id"), in)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Tab(res))
}

// ListTemplates handles GET /api/dashboard/templates.
func (h *Handlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	res, err := h.templates.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Templates(res))
}

// SaveTemplate handles POST /api/dashboard/templates.
func (h *Handlers) SaveTemplate(w http.ResponseWriter, r *http.Request) {
	var dto SaveTemplateDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.templates.SaveTab(r.Context(), dto.TabID, dto.Name)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Template(res))
}

// DeleteTemplate handles DELETE /api/dashboard/templates/{id}.
func (h *Handlers) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateTabFromTemplate handles POST /api/dashboard/templates/{id}/tabs.
func (h *Handlers) CreateTabFromTemplate(w http.ResponseWriter, r *http.Request) {
	var dto TabNameDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.templates.CreateTab(r.Context(), r.PathValue("id"), dto.Name)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Tab(res))
}
