package presenter

import (
	"time"

	"trisse-home/internal/modules/services/domain"
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
	// Container is the source container of a discovered tile, empty for
	// manual tiles; Hidden tiles lost their container (#78).
	Container string `json:"container"`
	Hidden    bool   `json:"hidden"`
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
		Container: v.Container, Hidden: v.Hidden,
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

// UptimeDTO summarises one service; percentages are null without data.
type UptimeDTO struct {
	ServiceID string     `json:"serviceId"`
	Day       *float64   `json:"day"`
	Week      *float64   `json:"week"`
	Month     *float64   `json:"month"`
	Quarter   *float64   `json:"quarter"`
	Days      []*float64 `json:"days"`
}

// Uptime maps the uptime summaries.
func Uptime(list []usecase.UptimeView) []UptimeDTO {
	out := make([]UptimeDTO, 0, len(list))
	for _, v := range list {
		out = append(out, UptimeDTO(v))
	}
	return out
}

// IncidentDTO is one outage.
type IncidentDTO struct {
	ID              string     `json:"id"`
	ServiceID       string     `json:"serviceId"`
	Start           time.Time  `json:"start"`
	End             *time.Time `json:"end"`
	DurationSeconds float64    `json:"durationSeconds"`
	Cause           string     `json:"cause"`
}

// Incidents maps incidents; open ones last until now.
func Incidents(list []domain.Incident, now time.Time) []IncidentDTO {
	out := make([]IncidentDTO, 0, len(list))
	for _, i := range list {
		dto := IncidentDTO{ID: i.ID, ServiceID: i.ServiceID, Start: i.Start, DurationSeconds: i.Duration(now).Seconds(), Cause: i.Cause}
		if !i.Open() {
			end := i.End
			dto.End = &end
		}
		out = append(out, dto)
	}
	return out
}
