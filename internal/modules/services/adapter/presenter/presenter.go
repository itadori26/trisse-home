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
	MAC           string    `json:"mac"`
	Position      int       `json:"position"`
	State         string    `json:"state" tstype:"'online' | 'offline' | 'unknown'"`
	HTTPStatus    int       `json:"httpStatus"`
	LatencyMs     int64     `json:"latencyMs"`
	CheckedAt     time.Time `json:"checkedAt"`
	Error         string    `json:"error"`
	// CertExpiresAt is the TLS certificate expiry (null for plain HTTP).
	CertExpiresAt   *time.Time `json:"certExpiresAt"`
	CertExpiresSoon bool       `json:"certExpiresSoon"`
}

// LatencyPointDTO is one point of a response time history.
type LatencyPointDTO struct {
	At        time.Time `json:"at"`
	LatencyMs float64   `json:"latencyMs"`
	Online    bool      `json:"online"`
}

// Service maps a ServiceView.
func Service(v usecase.ServiceView) ServiceDTO {
	dto := ServiceDTO{
		ID: v.ID, Name: v.Name, URL: v.URL, CheckURL: v.CheckURL, Icon: v.Icon, Description: v.Description,
		SkipTLSVerify: v.SkipTLSVerify, MAC: v.MAC, Position: v.Position, State: v.State, HTTPStatus: v.HTTPStatus,
		LatencyMs: v.LatencyMs, CheckedAt: v.CheckedAt, Error: v.Error, CertExpiresSoon: v.CertExpiresSoon,
	}
	if !v.CertExpiresAt.IsZero() {
		t := v.CertExpiresAt
		dto.CertExpiresAt = &t
	}
	return dto
}

// History maps the response time history per service ID.
func History(h map[string][]usecase.LatencyPointView) map[string][]LatencyPointDTO {
	out := make(map[string][]LatencyPointDTO, len(h))
	for id, ps := range h {
		list := make([]LatencyPointDTO, 0, len(ps))
		for _, p := range ps {
			list = append(list, LatencyPointDTO(p))
		}
		out[id] = list
	}
	return out
}

// Services maps a list of ServiceViews.
func Services(vs []usecase.ServiceView) []ServiceDTO {
	out := make([]ServiceDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Service(v))
	}
	return out
}
