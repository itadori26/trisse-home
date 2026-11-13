package presenter

import (
	"time"

	"trisse-home/internal/modules/docker/usecase"
)

// ContainerDTO is one container.
type ContainerDTO struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Image   string    `json:"image"`
	State   string    `json:"state"`
	Status  string    `json:"status"`
	Health  string    `json:"health" tstype:"'none' | 'healthy' | 'unhealthy' | 'starting'"`
	Project string    `json:"project"`
	Created time.Time `json:"created"`
}

// ListDTO is returned by GET /api/docker/containers.
type ListDTO struct {
	Containers []ContainerDTO `json:"containers"`
	Total      int            `json:"total"`
	Running    int            `json:"running"`
	Stopped    int            `json:"stopped"`
	Unhealthy  int            `json:"unhealthy"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	Error      string         `json:"error"`
}

// List maps a ListResponse.
func List(r usecase.ListResponse) ListDTO {
	dto := ListDTO{Total: r.Total, Running: r.Running, Stopped: r.Stopped, Unhealthy: r.Unhealthy, UpdatedAt: r.UpdatedAt, Error: r.Error,
		Containers: make([]ContainerDTO, 0, len(r.Containers))}
	// Labels are not part of the response: they can carry credentials
	// (e.g. basic auth hashes of a reverse proxy).
	for _, c := range r.Containers {
		dto.Containers = append(dto.Containers, ContainerDTO{ID: c.ID, Name: c.Name, Image: c.Image, State: c.State,
			Status: c.Status, Health: c.Health, Project: c.Project, Created: c.Created})
	}
	return dto
}
