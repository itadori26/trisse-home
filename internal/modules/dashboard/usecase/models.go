package usecase

// DashboardResponse lists all tabs in order.
type DashboardResponse struct {
	Tabs []TabResponse
}

// TabResponse describes one tab and its widgets.
type TabResponse struct {
	ID         string
	Name       string
	Position   int
	Background string
	Widgets    []WidgetResponse
}

// TemplateResponse describes a tab template.
type TemplateResponse struct {
	ID      string
	Name    string
	Widgets int
	Kinds   []string
}

// WidgetResponse describes one widget.
type WidgetResponse struct {
	ID     string
	TabID  string
	Kind   string
	X, Y   int
	W, H   int
	Config string
}

// CreateTabRequest is the input of ManageTabs.Create.
type CreateTabRequest struct {
	Name string
}

// RenameTabRequest is the input of ManageTabs.Rename.
type RenameTabRequest struct {
	TabID string
	Name  string
}

// ReorderTabsRequest lists all tab IDs in their new order.
type ReorderTabsRequest struct {
	TabIDs []string
}

// AddWidgetRequest is the input of ManageWidgets.Add.
type AddWidgetRequest struct {
	TabID  string
	Kind   string
	X, Y   int
	W, H   int
	Config string
}

// ConfigureWidgetRequest is the input of ManageWidgets.Configure.
type ConfigureWidgetRequest struct {
	WidgetID string
	Config   string
}

// WidgetPlacement is the new rectangle of one widget.
type WidgetPlacement struct {
	WidgetID string
	X, Y     int
	W, H     int
}

// SaveLayoutRequest stores the positions of the widgets of one tab.
type SaveLayoutRequest struct {
	TabID      string
	Placements []WidgetPlacement
}
