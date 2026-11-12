package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"trisse-home/internal/modules/arr/domain"
)

const maxResponseBytes = 4 << 20

// client is the HTTP part shared by all *arr apps: same auth header, same
// error shapes; only the API version and the payloads differ.
type client struct {
	http    *http.Client
	version string // "v3" or "v1"
}

func (c client) get(ctx context.Context, in domain.Instance, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, in.URL+"/api/"+c.version+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Api-Key", in.APIKey)
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return fmt.Errorf("the API key was rejected (%s)", res.Status)
	default:
		return fmt.Errorf("unexpected answer: %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response: %w", err)
	}
	return nil
}

// check reads system/status and verifies the app is the expected one, so
// a Radarr URL saved as Sonarr is reported instead of failing later.
func (c client) check(ctx context.Context, in domain.Instance) error {
	var status struct {
		AppName string `json:"appName"`
	}
	if err := c.get(ctx, in, "/system/status", &status); err != nil {
		return err
	}
	if status.AppName != "" && !strings.EqualFold(status.AppName, string(in.App)) {
		return fmt.Errorf("this is %s, not %s", status.AppName, in.App.Title())
	}
	return nil
}

type healthPayload []struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c client) health(ctx context.Context, in domain.Instance) ([]domain.HealthIssue, error) {
	var payload healthPayload
	if err := c.get(ctx, in, "/health", &payload); err != nil {
		return nil, err
	}
	out := make([]domain.HealthIssue, 0, len(payload))
	for _, h := range payload {
		out = append(out, domain.HealthIssue{Instance: in.Name, Severity: domain.Severity(strings.ToLower(h.Type)), Message: h.Message})
	}
	return out, nil
}
