package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"trisse-home/internal/modules/system/usecase"
)

// DefaultReleasesURL is the GitHub API endpoint of the latest release.
const DefaultReleasesURL = "https://api.github.com/repos/itadori26/trisse-home/releases/latest"

// GitHubReleases implements usecase.ReleaseSource.
type GitHubReleases struct {
	client *http.Client
	url    string
}

// NewGitHubReleases returns a release source for url.
func NewGitHubReleases(client *http.Client, url string) *GitHubReleases {
	return &GitHubReleases{client: client, url: url}
}

// Latest implements usecase.ReleaseSource; ok is false when no release exists yet.
func (g *GitHubReleases) Latest(ctx context.Context) (usecase.Release, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)
	if err != nil {
		return usecase.Release{}, false, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "trisse-home")
	res, err := g.client.Do(req)
	if err != nil {
		return usecase.Release{}, false, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return usecase.Release{}, false, nil
	}
	if res.StatusCode != http.StatusOK {
		return usecase.Release{}, false, fmt.Errorf("GitHub answered %s", res.Status)
	}
	var body struct {
		TagName     string    `json:"tag_name"`
		HTMLURL     string    `json:"html_url"`
		PublishedAt time.Time `json:"published_at"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return usecase.Release{}, false, fmt.Errorf("decode release: %w", err)
	}
	return usecase.Release{Version: body.TagName, URL: body.HTMLURL, PublishedAt: body.PublishedAt}, true, nil
}

// UpdateSettings implements usecase.UpdateSettings in SQLite.
type UpdateSettings struct{ db *sql.DB }

// NewUpdateSettings returns the settings store.
func NewUpdateSettings(db *sql.DB) *UpdateSettings { return &UpdateSettings{db: db} }

// CheckEnabled implements usecase.UpdateSettings (default: off).
func (s *UpdateSettings) CheckEnabled(ctx context.Context) (bool, error) {
	var on int
	err := s.db.QueryRowContext(ctx, `SELECT update_check FROM system_settings WHERE id = 1`).Scan(&on)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return on == 1, err
}

// SetCheckEnabled implements usecase.UpdateSettings.
func (s *UpdateSettings) SetCheckEnabled(ctx context.Context, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO system_settings (id, update_check) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET update_check = excluded.update_check`, v)
	return err
}
