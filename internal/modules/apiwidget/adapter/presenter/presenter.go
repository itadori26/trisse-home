// Package presenter maps endpoints and values to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/modules/apiwidget/usecase"
	"trisse-home/internal/shared/kernel"
)

// ExtractionDTO is one configured value.
type ExtractionDTO struct {
	Label string   `json:"label"`
	Path  string   `json:"path"`
	Unit  string   `json:"unit"`
	Warn  *float64 `json:"warn"`
	Crit  *float64 `json:"crit"`
	Below bool     `json:"below"`
}

// EndpointDTO is an endpoint; the header value is never returned.
type EndpointDTO struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	URL             string          `json:"url"`
	HeaderName      string          `json:"headerName"`
	HasHeader       bool            `json:"hasHeader"`
	IntervalSeconds int             `json:"intervalSeconds"`
	AllowPublic     bool            `json:"allowPublic"`
	Extractions     []ExtractionDTO `json:"extractions"`
}

// ValueDTO is one extracted value.
type ValueDTO struct {
	Label  string   `json:"label"`
	Unit   string   `json:"unit"`
	Text   string   `json:"text"`
	Number *float64 `json:"number"`
	// Status is ok, warning, critical or unknown.
	Status string `json:"status"`
	// Error says why the path matched nothing (technical, for building paths).
	Error string `json:"error"`
}

// ValuesDTO is returned by GET /api/apiwidget/endpoints/{id}/values.
type ValuesDTO struct {
	// Name of the endpoint, shown as the widget heading.
	Name      string     `json:"name"`
	Values    []ValueDTO `json:"values"`
	FetchedAt time.Time  `json:"fetchedAt"`
	// Error is the English message of a failed request; ErrorKey and
	// ErrorParams let the client show it translated.
	Error       string            `json:"error"`
	ErrorKey    string            `json:"errorKey"`
	ErrorParams map[string]string `json:"errorParams"`
}

// PreviewDTO is returned by POST /api/apiwidget/preview.
type PreviewDTO struct {
	ValuesDTO `tstype:",extends"`
	// Body is the beginning of the response (at most 4 KB).
	Body string `json:"body"`
}

// Endpoint maps an endpoint.
func Endpoint(e domain.Endpoint) EndpointDTO {
	dto := EndpointDTO{ID: e.ID, Name: e.Name, URL: e.URL, HeaderName: e.HeaderName, HasHeader: e.HasHeader(),
		IntervalSeconds: int(e.Interval / time.Second), AllowPublic: e.AllowPublic, Extractions: make([]ExtractionDTO, 0, len(e.Extractions))}
	for _, x := range e.Extractions {
		dto.Extractions = append(dto.Extractions, ExtractionDTO(x))
	}
	return dto
}

// Endpoints maps a list.
func Endpoints(es []domain.Endpoint) []EndpointDTO {
	out := make([]EndpointDTO, 0, len(es))
	for _, e := range es {
		out = append(out, Endpoint(e))
	}
	return out
}

// Values maps the values of a fetch.
func Values(v usecase.Values) ValuesDTO {
	dto := ValuesDTO{Name: v.Name, FetchedAt: v.FetchedAt, Values: make([]ValueDTO, 0, len(v.Values))}
	for _, x := range v.Values {
		dto.Values = append(dto.Values, ValueDTO{Label: x.Label, Unit: x.Unit, Text: x.Text, Number: x.Number, Status: string(x.Status), Error: x.Error})
	}
	if v.Error != nil {
		dto.Error, dto.ErrorKey, dto.ErrorParams = v.Error.Message, kernel.MessageKey(v.Error.Template), v.Error.Params
	}
	return dto
}

// Preview maps a preview.
func Preview(p usecase.Preview) PreviewDTO {
	return PreviewDTO{ValuesDTO: Values(p.Values), Body: p.Body}
}
