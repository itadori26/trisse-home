package domain

import (
	"net/url"
	"regexp"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// ErrNotConfigured is returned while Immich is not configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Immich is not configured yet.")

// ErrThumbnailNotFound is returned for an asset that is not part of today's memories.
var ErrThumbnailNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The thumbnail does not exist.")

// Settings are the connection details of Immich.
type Settings struct {
	URL    string
	APIKey string
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, apiKey string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like http://192.168.1.2:2283.")
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return Settings{}, kernel.NewValidationError("apiKey", "The API key is required.")
	}
	return Settings{URL: rawURL, APIKey: apiKey}, nil
}

// assetIDPattern matches Immich asset IDs (UUIDs).
var assetIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidAssetID reports whether id is an Immich asset ID.
func ValidAssetID(id string) bool { return assetIDPattern.MatchString(id) }

// LibraryStats are the numbers of the library and the server's disk.
type LibraryStats struct {
	Photos, Videos      int
	DiskUsed, DiskTotal int64
}

// DiskPercent returns the used share of the disk (0 without a size).
func (s LibraryStats) DiskPercent() float64 {
	if s.DiskTotal <= 0 {
		return 0
	}
	return float64(s.DiskUsed) / float64(s.DiskTotal) * 100
}

// Memory is "on this day" a number of years ago.
type Memory struct {
	YearsAgo int
	// AssetIDs are the photos to show, at most MaxMemoryAssets.
	AssetIDs []string
}

// MaxMemoryAssets bounds the thumbnails per memory.
const MaxMemoryAssets = 6

// Overview is what the widget shows.
type Overview struct {
	Version  string
	Stats    LibraryStats
	Memories []Memory
}

// Thumbnail is a small preview image.
type Thumbnail struct {
	ContentType string
	Data        []byte
}
