package controller

// ExtractionInputDTO is one value to pick out of the response.
type ExtractionInputDTO struct {
	Label string   `json:"label"`
	Path  string   `json:"path"`
	Unit  string   `json:"unit"`
	Warn  *float64 `json:"warn"`
	Crit  *float64 `json:"crit"`
	Below bool     `json:"below"`
}

// EndpointInputDTO is the body of POST/PUT /api/apiwidget/endpoints and of
// the preview. An empty headerValue with keepHeader keeps the stored value.
type EndpointInputDTO struct {
	Name            string               `json:"name"`
	URL             string               `json:"url"`
	HeaderName      string               `json:"headerName"`
	HeaderValue     string               `json:"headerValue"`
	KeepHeader      bool                 `json:"keepHeader"`
	IntervalSeconds int                  `json:"intervalSeconds"`
	AllowPublic     bool                 `json:"allowPublic"`
	Extractions     []ExtractionInputDTO `json:"extractions"`
}

// PreviewInputDTO is the body of POST /api/apiwidget/preview; id is the
// endpoint being edited (its stored header value can be kept) or empty.
type PreviewInputDTO struct {
	ID       string           `json:"id"`
	Endpoint EndpointInputDTO `json:"endpoint"`
}
