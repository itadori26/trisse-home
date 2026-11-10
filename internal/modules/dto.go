package modules

// InfoResponse is the DTO returned by GET /api/modules.
type InfoResponse struct {
	Modules     []string `json:"modules"`
	WidgetKinds []string `json:"widgetKinds"`
}
