package controller

import "encoding/json"

// TabNameDTO is the body of POST /api/dashboard/tabs and PATCH /api/dashboard/tabs/{id}.
type TabNameDTO struct {
	Name string `json:"name"`
}

// TabOrderDTO is the body of PUT /api/dashboard/tabs/order.
type TabOrderDTO struct {
	TabIDs []string `json:"tabIds"`
}

// AddWidgetDTO is the body of POST /api/dashboard/tabs/{id}/widgets.
type AddWidgetDTO struct {
	Kind   string          `json:"kind"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	W      int             `json:"w"`
	H      int             `json:"h"`
	Config json.RawMessage `json:"config,omitempty" tstype:"{ [key: string]: unknown }"`
}

// WidgetConfigDTO is the body of PATCH /api/dashboard/widgets/{id}.
type WidgetConfigDTO struct {
	Config json.RawMessage `json:"config" tstype:"{ [key: string]: unknown }"`
}

// PlacementDTO is the rectangle of one widget.
type PlacementDTO struct {
	ID string `json:"id"`
	X  int    `json:"x"`
	Y  int    `json:"y"`
	W  int    `json:"w"`
	H  int    `json:"h"`
}

// LayoutDTO is the body of PUT /api/dashboard/tabs/{id}/layout.
type LayoutDTO struct {
	Items []PlacementDTO `json:"items"`
}

// TabBackgroundDTO is the body of PUT /api/dashboard/tabs/{id}/background;
// an empty background uses the global one.
type TabBackgroundDTO struct {
	Background string `json:"background"`
}

// TabRuleDTO is one default-tab rule (#71).
type TabRuleDTO struct {
	Days     []int    `json:"days"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Networks []string `json:"networks"`
}

// TabRulesDTO is the body of PUT /api/dashboard/tabs/{id}/rules.
type TabRulesDTO struct {
	Rules []TabRuleDTO `json:"rules"`
}

// MobileItemInputDTO places a widget in the phone layout (#92).
type MobileItemInputDTO struct {
	ID     string `json:"id"`
	Hidden bool   `json:"hidden"`
	Rows   int    `json:"rows"`
}

// MobileLayoutDTO is the body of PUT /api/dashboard/tabs/{id}/mobile-layout.
type MobileLayoutDTO struct {
	Items []MobileItemInputDTO `json:"items"`
}

// SaveTemplateDTO is the body of POST /api/dashboard/templates.
type SaveTemplateDTO struct {
	TabID string `json:"tabId"`
	Name  string `json:"name"`
}
