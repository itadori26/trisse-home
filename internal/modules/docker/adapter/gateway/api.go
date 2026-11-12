package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"trisse-home/internal/modules/docker/domain"
)

const maxResponseBytes = 4 << 20

// API lists containers via the Docker Engine HTTP API.
type API struct {
	client  *http.Client
	baseURL string
}

// NewAPI returns a client for baseURL, e.g. http://docker-socket-proxy:2375.
func NewAPI(client *http.Client, baseURL string) *API {
	return &API{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

// containerPayload is the subset of GET /containers/json we use.
type containerPayload struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Created int64             `json:"Created"`
	Labels  map[string]string `json:"Labels"`
}

// Containers implements usecase.ContainerSource.
func (a *API) Containers(ctx context.Context) ([]domain.Container, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/containers/json?all=1", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docker api: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker api returned %s", res.Status)
	}
	var payload []containerPayload
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode docker response: %w", err)
	}
	out := make([]domain.Container, 0, len(payload))
	for _, p := range payload {
		name := p.ID
		if len(p.Names) > 0 {
			name = strings.TrimPrefix(p.Names[0], "/")
		}
		id := p.ID
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, domain.Container{
			ID: id, Name: name, Image: p.Image, State: p.State, Status: p.Status,
			Project: p.Labels["com.docker.compose.project"], Created: time.Unix(p.Created, 0).UTC(), Labels: p.Labels,
		})
	}
	return out, nil
}
