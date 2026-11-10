// Package presenter maps dashboard use case responses to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"encoding/json"

	"trisse-home/internal/modules/dashboard/usecase"
)

// DashboardDTO is returned by GET /api/dashboard.
type DashboardDTO struct {
	Tabs []TabDTO `json:"tabs"`
}

// TabDTO describes a tab with its widgets.
type TabDTO struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Position int         `json:"position"`
	Widgets  []WidgetDTO `json:"widgets"`
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
	out := DashboardDTO{Tabs: make([]TabDTO, 0, len(r.Tabs))}
	for _, t := range r.Tabs {
		out.Tabs = append(out.Tabs, Tab(t))
	}
	return out
}

// Tab maps a TabResponse.
func Tab(t usecase.TabResponse) TabDTO {
	dto := TabDTO{ID: t.ID, Name: t.Name, Position: t.Position, Widgets: make([]WidgetDTO, 0, len(t.Widgets))}
	for _, w := range t.Widgets {
		dto.Widgets = append(dto.Widgets, Widget(w))
	}
	return dto
}

// Widget maps a WidgetResponse.
func Widget(w usecase.WidgetResponse) WidgetDTO {
	return WidgetDTO{ID: w.ID, TabID: w.TabID, Kind: w.Kind, X: w.X, Y: w.Y, W: w.W, H: w.H, Config: json.RawMessage(w.Config)}
}
