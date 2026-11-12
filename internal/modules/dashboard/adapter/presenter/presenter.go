package presenter

import (
	"encoding/json"

	"trisse-home/internal/modules/dashboard/usecase"
)

// DashboardDTO is returned by GET /api/dashboard.
type DashboardDTO struct {
	Tabs []TabDTO `json:"tabs"`
	// DefaultTabID is the tab whose rules match now ("" = keep the last used one, #71).
	DefaultTabID string `json:"defaultTabId"`
}

// RuleDTO opens a tab by default on weekdays (0 = Sunday), in a time range
// ("HH:MM", empty = whole day) and/or from networks (CIDR).
type RuleDTO struct {
	Days     []int    `json:"days"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Networks []string `json:"networks"`
}

// TabDTO describes a tab with its widgets.
type TabDTO struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Position   int         `json:"position"`
	Background string      `json:"background"`
	Rules      []RuleDTO   `json:"rules"`
	Widgets    []WidgetDTO `json:"widgets"`
}

// WidgetDTO describes a widget. Config is the widget's JSON configuration.
type WidgetDTO struct {
	ID     string          `json:"id"`
	TabID  string          `json:"tabId"`
	Kind   string          `json:"kind"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	W      int             `json:"w"`
	H      int             `json:"h"`
	Config json.RawMessage `json:"config" tstype:"{ [key: string]: unknown }"`
}

// Dashboard maps a DashboardResponse.
func Dashboard(r usecase.DashboardResponse) DashboardDTO {
	out := DashboardDTO{Tabs: make([]TabDTO, 0, len(r.Tabs)), DefaultTabID: r.DefaultTabID}
	for _, t := range r.Tabs {
		out.Tabs = append(out.Tabs, Tab(t))
	}
	return out
}

// Tab maps a TabResponse.
func Tab(t usecase.TabResponse) TabDTO {
	dto := TabDTO{ID: t.ID, Name: t.Name, Position: t.Position, Background: t.Background, Widgets: make([]WidgetDTO, 0, len(t.Widgets)),
		Rules: make([]RuleDTO, 0, len(t.Rules))}
	for _, r := range t.Rules {
		dto.Rules = append(dto.Rules, RuleDTO{Days: append([]int{}, r.Days...), From: r.From, To: r.To, Networks: append([]string{}, r.Networks...)})
	}
	for _, w := range t.Widgets {
		dto.Widgets = append(dto.Widgets, Widget(w))
	}
	return dto
}

// Widget maps a WidgetResponse.
func Widget(w usecase.WidgetResponse) WidgetDTO {
	return WidgetDTO{ID: w.ID, TabID: w.TabID, Kind: w.Kind, X: w.X, Y: w.Y, W: w.W, H: w.H, Config: json.RawMessage(w.Config)}
}

// TemplateDTO describes a tab template.
type TemplateDTO struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Widgets int      `json:"widgets"`
	Kinds   []string `json:"kinds"`
}

// Template maps a TemplateResponse.
func Template(t usecase.TemplateResponse) TemplateDTO { return TemplateDTO(t) }

// Templates maps a list of templates.
func Templates(ts []usecase.TemplateResponse) []TemplateDTO {
	out := make([]TemplateDTO, 0, len(ts))
	for _, t := range ts {
		out = append(out, Template(t))
	}
	return out
}
