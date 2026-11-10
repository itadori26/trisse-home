package presenter

import (
	"time"

	"trisse-home/internal/modules/services/usecase"
)

// ServiceDTO is a service tile with its status.
type ServiceDTO struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	CheckURL      string    `json:"checkUrl"`
	Icon          string    `json:"icon"`
	Description   string    `json:"description"`
	SkipTLSVerify bool      `json:"skipTlsVerify"`
	Position      int       `json:"position"`
	State         string    `json:"state" tstype:"'online' | 'offline' | 'unknown'"`
	HTTPStatus    int       `json:"httpStatus"`
	LatencyMs     int64     `json:"latencyMs"`
	CheckedAt     time.Time `json:"checkedAt"`
	Error         string    `json:"error"`
}

// Service maps a ServiceView.
func Service(v usecase.ServiceView) ServiceDTO { return ServiceDTO(v) }

// Services maps a list of ServiceViews.
func Services(vs []usecase.ServiceView) []ServiceDTO {
	out := make([]ServiceDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Service(v))
	}
	return out
}
