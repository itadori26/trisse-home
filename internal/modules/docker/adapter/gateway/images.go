package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"trisse-home/internal/modules/docker/usecase"
)

// RunningImages implements usecase.LocalImageReader. It needs IMAGES=1 on
// the read-only socket proxy (GET /images/{id}/json only).
func (a *API) RunningImages(ctx context.Context) ([]usecase.LocalImage, error) {
	var containers []struct {
		Names   []string `json:"Names"`
		Image   string   `json:"Image"`
		ImageID string   `json:"ImageID"`
	}
	if err := a.getJSON(ctx, "/containers/json", &containers); err != nil {
		return nil, err
	}
	byRef := map[string]*usecase.LocalImage{}
	imageOf := map[string]string{}
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		img, ok := byRef[c.Image]
		if !ok {
			img = &usecase.LocalImage{Ref: c.Image}
			byRef[c.Image] = img
			imageOf[c.Image] = c.ImageID
		}
		img.Containers = append(img.Containers, name)
	}
	out := make([]usecase.LocalImage, 0, len(byRef))
	for ref, img := range byRef {
		var inspect struct {
			RepoDigests []string `json:"RepoDigests"`
			Config      struct {
				Labels map[string]string `json:"Labels"`
			} `json:"Config"`
		}
		if err := a.getJSON(ctx, "/images/"+url.PathEscape(imageOf[ref])+"/json", &inspect); err != nil {
			return nil, err
		}
		img.RepoDigests, img.Labels = inspect.RepoDigests, inspect.Config.Labels
		sort.Strings(img.Containers)
		out = append(out, *img)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ref < out[j].Ref })
	return out, nil
}

func (a *API) getJSON(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("docker api: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("docker api returned %s for %s", res.Status, path)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("decode docker response: %w", err)
	}
	return nil
}
