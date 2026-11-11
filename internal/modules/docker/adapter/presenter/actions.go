package presenter

import "trisse-home/internal/modules/docker/usecase"

// CapabilitiesDTO is returned by GET /api/docker/actions.
type CapabilitiesDTO struct {
	Enabled    bool     `json:"enabled"`
	Containers []string `json:"containers"`
}

// LogsDTO is returned by GET /api/docker/containers/{name}/logs.
type LogsDTO struct {
	Lines []string `json:"lines"`
}

// Capabilities maps the action capabilities.
func Capabilities(c usecase.Capabilities) CapabilitiesDTO {
	return CapabilitiesDTO(c)
}
