package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/immich/domain"
)

const (
	maxJSONBytes      = 2 << 20
	maxThumbnailBytes = 1 << 20
)

// API implements usecase.Client with an Immich API key. The documented key
// needs only asset.statistics, server.storage, memory.read and asset.view.
type API struct{ client *http.Client }

// NewAPI returns a client.
func NewAPI(client *http.Client) *API { return &API{client: client} }

// Version implements usecase.Client (public endpoint).
func (a *API) Version(ctx context.Context, s domain.Settings) (string, error) {
	var v struct{ Major, Minor, Patch int }
	if err := a.getJSON(ctx, s, "/api/server/version", &v); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch), nil
}

// Stats implements usecase.Client: the key owner's assets and the disk.
func (a *API) Stats(ctx context.Context, s domain.Settings) (domain.LibraryStats, error) {
	var assets struct {
		Images int `json:"images"`
		Videos int `json:"videos"`
	}
	if err := a.getJSON(ctx, s, "/api/assets/statistics", &assets); err != nil {
		return domain.LibraryStats{}, err
	}
	out := domain.LibraryStats{Photos: assets.Images, Videos: assets.Videos}
	var storage struct {
		DiskSizeRaw int64 `json:"diskSizeRaw"`
		DiskUseRaw  int64 `json:"diskUseRaw"`
	}
	// The disk is optional: a key without server.storage still shows the library.
	if err := a.getJSON(ctx, s, "/api/server/storage", &storage); err == nil {
		out.DiskUsed, out.DiskTotal = storage.DiskUseRaw, storage.DiskSizeRaw
	}
	return out, nil
}

// Memories implements usecase.Client.
func (a *API) Memories(ctx context.Context, s domain.Settings, day time.Time) ([]domain.Memory, error) {
	var payload []struct {
		Type string `json:"type"`
		Data struct {
			Year int `json:"year"`
		} `json:"data"`
		Assets []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"assets"`
	}
	q := url.Values{"for": {day.Format(time.RFC3339)}}
	if err := a.getJSON(ctx, s, "/api/memories?"+q.Encode(), &payload); err != nil {
		return nil, err
	}
	var out []domain.Memory
	for _, m := range payload {
		if m.Type != "on_this_day" || m.Data.Year <= 0 {
			continue
		}
		mem := domain.Memory{YearsAgo: day.Year() - m.Data.Year}
		for _, asset := range m.Assets {
			if asset.Type == "IMAGE" && domain.ValidAssetID(asset.ID) {
				mem.AssetIDs = append(mem.AssetIDs, asset.ID)
			}
		}
		if mem.YearsAgo > 0 && len(mem.AssetIDs) > 0 {
			out = append(out, mem)
		}
	}
	return out, nil
}

// Thumbnail implements usecase.Client.
func (a *API) Thumbnail(ctx context.Context, s domain.Settings, assetID string) (domain.Thumbnail, error) {
	res, err := a.get(ctx, s, "/api/assets/"+url.PathEscape(assetID)+"/thumbnail?size=thumbnail")
	if err != nil {
		return domain.Thumbnail{}, err
	}
	defer res.Body.Close()
	ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if !strings.HasPrefix(ct, "image/") || ct == "image/svg+xml" {
		return domain.Thumbnail{}, fmt.Errorf("not a raster image: %q", ct)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxThumbnailBytes+1))
	if err != nil {
		return domain.Thumbnail{}, fmt.Errorf("read thumbnail: %w", err)
	}
	if len(data) > maxThumbnailBytes {
		return domain.Thumbnail{}, fmt.Errorf("thumbnail larger than %s bytes", strconv.Itoa(maxThumbnailBytes))
	}
	return domain.Thumbnail{ContentType: ct, Data: data}, nil
}

func (a *API) getJSON(ctx context.Context, s domain.Settings, path string, dst any) error {
	res, err := a.get(ctx, s, path)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(io.LimitReader(res.Body, maxJSONBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from Immich: %w", err)
	}
	return nil
}

func (a *API) get(ctx context.Context, s domain.Settings, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-api-key", s.APIKey)
	req.Header.Set("Accept", "application/json")
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	switch res.StatusCode {
	case http.StatusOK:
		return res, nil
	case http.StatusUnauthorized:
		_ = res.Body.Close()
		return nil, fmt.Errorf("the API key was rejected (%s)", res.Status)
	case http.StatusForbidden:
		_ = res.Body.Close()
		return nil, fmt.Errorf("the API key lacks a permission (%s)", res.Status)
	default:
		_ = res.Body.Close()
		return nil, fmt.Errorf("unexpected answer from Immich: %s", res.Status)
	}
}
