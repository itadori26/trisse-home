package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"trisse-home/internal/modules/media/domain"
)

const (
	maxJSONBytes   = 4 << 20
	maxPosterBytes = 2 << 20
)

// getJSON requests path with the provider's auth header and decodes JSON.
func getJSON(ctx context.Context, client *http.Client, s domain.Settings, path string, auth func(*http.Request), dst any) error {
	res, err := get(ctx, client, s.URL+path, auth)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(io.LimitReader(res.Body, maxJSONBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from the media server: %w", err)
	}
	return nil
}

// getImage downloads a poster; anything but a small image is rejected.
func getImage(ctx context.Context, client *http.Client, url string, auth func(*http.Request)) (domain.Poster, error) {
	res, err := get(ctx, client, url, auth)
	if err != nil {
		return domain.Poster{}, err
	}
	defer res.Body.Close()
	ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if !strings.HasPrefix(ct, "image/") || ct == "image/svg+xml" {
		return domain.Poster{}, fmt.Errorf("not a raster image: %q", ct)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxPosterBytes+1))
	if err != nil {
		return domain.Poster{}, fmt.Errorf("read poster: %w", err)
	}
	if len(data) > maxPosterBytes {
		return domain.Poster{}, fmt.Errorf("poster larger than %d bytes", maxPosterBytes)
	}
	return domain.Poster{ContentType: ct, Data: data}, nil
}

func get(ctx context.Context, client *http.Client, url string, auth func(*http.Request)) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	auth(req)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	switch res.StatusCode {
	case http.StatusOK:
		return res, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		_ = res.Body.Close()
		return nil, fmt.Errorf("the API key or token was rejected (%s)", res.Status)
	default:
		_ = res.Body.Close()
		return nil, fmt.Errorf("unexpected answer from the media server: %s", res.Status)
	}
}
