package main

import (
	"context"

	"trisse-home/internal/modules/docker"
	servicesdomain "trisse-home/internal/modules/services/domain"
)

// containerLabels adapts the docker module's read model to the container
// source of the services module (Docker autodiscovery, #78): modules never
// import each other, the composition root translates.
type containerLabels struct{ docker *docker.Module }

// Containers implements services/usecase.ContainerSource.
func (c containerLabels) Containers(ctx context.Context) ([]servicesdomain.Container, error) {
	list, err := c.docker.Containers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]servicesdomain.Container, 0, len(list))
	for _, v := range list {
		out = append(out, servicesdomain.Container{Name: v.Name, Labels: v.Labels})
	}
	return out, nil
}
