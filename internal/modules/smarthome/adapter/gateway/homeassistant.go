package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"trisse-home/internal/modules/smarthome/domain"
)

// maxStatesBytes bounds GET /api/states; large installations have a few
// thousand entities.
const maxStatesBytes = 16 << 20

// HomeAssistant implements usecase.Provider with the Home Assistant REST
// API and a long-lived access token.
type HomeAssistant struct{ client *http.Client }

// NewHomeAssistant returns a client.
func NewHomeAssistant(client *http.Client) *HomeAssistant { return &HomeAssistant{client: client} }

type haState struct {
	EntityID   string `json:"entity_id"`
	State      string `json:"state"`
	Attributes struct {
		FriendlyName string `json:"friendly_name"`
		Unit         string `json:"unit_of_measurement"`
	} `json:"attributes"`
	LastChanged time.Time `json:"last_changed"`
}

// States implements usecase.Provider.
func (h *HomeAssistant) States(ctx context.Context, s domain.Settings) ([]domain.Entity, error) {
	var payload []haState
	if err := h.do(ctx, s, http.MethodGet, "/api/states", nil, &payload); err != nil {
		return nil, err
	}
	out := make([]domain.Entity, 0, len(payload))
	for _, p := range payload {
		if !domain.ValidEntityID(p.EntityID) {
			continue
		}
		name := p.Attributes.FriendlyName
		if name == "" {
			name = p.EntityID
		}
		out = append(out, domain.Entity{ID: p.EntityID, Name: name, State: p.State, Unit: p.Attributes.Unit, LastChanged: p.LastChanged})
	}
	return out, nil
}

// Call implements usecase.Provider: POST /api/services/<domain>/<service>.
func (h *HomeAssistant) Call(ctx context.Context, s domain.Settings, svc domain.Service, entityID string) error {
	body, err := json.Marshal(map[string]string{"entity_id": entityID})
	if err != nil {
		return fmt.Errorf("encode service call: %w", err)
	}
	path := "/api/services/" + url.PathEscape(svc.Domain) + "/" + url.PathEscape(svc.Name)
	var changed []json.RawMessage
	return h.do(ctx, s, http.MethodPost, path, body, &changed)
}

func (h *HomeAssistant) do(ctx context.Context, s domain.Settings, method, path string, body []byte, dst any) error {
	req, err := http.NewRequestWithContext(ctx, method, s.URL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return fmt.Errorf("the access token was rejected (%s)", res.Status)
	default:
		return fmt.Errorf("unexpected answer from Home Assistant: %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxStatesBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from Home Assistant: %w", err)
	}
	return nil
}
