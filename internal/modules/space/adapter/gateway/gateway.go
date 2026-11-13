// Package gateway stores the space settings and talks to Launch Library 2
// and the NASA APOD API.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/space/domain"
	"trisse-home/internal/modules/space/usecase"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are exported without the encrypted key.
var TransferTables = []sqlutil.Table{{Name: "space_settings", Exclude: []string{"nasa_key_encrypted"}}}

// Default endpoints.
const (
	DefaultLaunchURL  = "https://ll.thespacedevs.com/2.2.0/launch/upcoming/"
	DefaultPictureURL = "https://api.nasa.gov/planetary/apod"
	demoKey           = "DEMO_KEY"
	maxJSONBytes      = 2 << 20
	// MaxImageBytes bounds the proxied picture.
	MaxImageBytes = 15 << 20
)

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// SettingsRepository implements usecase.SettingsRepository.
type SettingsRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB, cipher Cipher) *SettingsRepository {
	return &SettingsRepository{db: db, cipher: cipher}
}

// Load implements usecase.SettingsRepository; no row means everything off.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var enc string
	err := r.db.QueryRowContext(ctx, `SELECT launches, picture, nasa_key_encrypted FROM space_settings WHERE id = 1`).Scan(&s.Launches, &s.Picture, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if enc != "" {
		if s.NASAKey, err = r.cipher.Decrypt(enc); err != nil {
			return domain.Settings{}, fmt.Errorf("decrypt NASA key (was SECRET_KEY changed?): %w", err)
		}
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc := ""
	if s.NASAKey != "" {
		var err error
		if enc, err = r.cipher.Encrypt(s.NASAKey); err != nil {
			return fmt.Errorf("encrypt NASA key: %w", err)
		}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO space_settings (id, launches, picture, nasa_key_encrypted) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET launches = excluded.launches, picture = excluded.picture, nasa_key_encrypted = excluded.nasa_key_encrypted`,
		s.Launches, s.Picture, enc)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

func getJSON(ctx context.Context, client *http.Client, u string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server answered %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxJSONBytes)).Decode(dst); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

// LaunchLibrary implements usecase.LaunchProvider.
type LaunchLibrary struct {
	client *http.Client
	base   string
}

// NewLaunchLibrary returns a Launch Library 2 client.
func NewLaunchLibrary(client *http.Client, base string) *LaunchLibrary {
	return &LaunchLibrary{client: client, base: base}
}

type ll2Response struct {
	Results []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Net    string `json:"net"`
		Status struct {
			Abbrev string `json:"abbrev"`
		} `json:"status"`
		Rocket struct {
			Configuration struct {
				Name string `json:"name"`
			} `json:"configuration"`
		} `json:"rocket"`
		Mission *struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"mission"`
		Provider struct {
			Name string `json:"name"`
		} `json:"launch_service_provider"`
		Pad struct {
			Name     string `json:"name"`
			Location struct {
				Name string `json:"name"`
			} `json:"location"`
		} `json:"pad"`
		WebcastLive bool `json:"webcast_live"`
	} `json:"results"`
}

// Upcoming implements usecase.LaunchProvider.
func (l *LaunchLibrary) Upcoming(ctx context.Context, limit int) ([]domain.Launch, error) {
	q := url.Values{"limit": {strconv.Itoa(limit)}, "mode": {"normal"}, "hide_recent_previous": {"true"}}
	var res ll2Response
	if err := getJSON(ctx, l.client, l.base+"?"+q.Encode(), &res); err != nil {
		return nil, err
	}
	out := make([]domain.Launch, 0, len(res.Results))
	for _, r := range res.Results {
		net, err := time.Parse(time.RFC3339, r.Net)
		if err != nil {
			continue
		}
		launch := domain.Launch{ID: r.ID, Rocket: r.Rocket.Configuration.Name, Provider: r.Provider.Name, Pad: r.Pad.Name,
			Location: r.Pad.Location.Name, Net: net, Status: domain.ParseLaunchStatus(r.Status.Abbrev), WebcastLive: r.WebcastLive}
		// The name is "Rocket | Mission"; the mission object has more details.
		launch.Mission = r.Name
		if _, mission, ok := strings.Cut(r.Name, " | "); ok {
			launch.Mission = mission
		}
		if r.Mission != nil {
			if r.Mission.Name != "" {
				launch.Mission = r.Mission.Name
			}
			launch.MissionType = r.Mission.Type
			launch.Description = domain.Truncate(r.Mission.Description, domain.MaxDescription)
		}
		out = append(out, launch)
	}
	return out, nil
}

// APOD implements usecase.PictureProvider.
type APOD struct {
	client *http.Client
	base   string
}

// NewAPOD returns a NASA APOD client.
func NewAPOD(client *http.Client, base string) *APOD { return &APOD{client: client, base: base} }

type apodResponse struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
	Copyright   string `json:"copyright"`
	MediaType   string `json:"media_type"`
	URL         string `json:"url"`
	HDURL       string `json:"hdurl"`
}

// Today implements usecase.PictureProvider.
func (a *APOD) Today(ctx context.Context, apiKey string) (domain.Picture, error) {
	if apiKey == "" {
		apiKey = demoKey
	}
	var r apodResponse
	if err := getJSON(ctx, a.client, a.base+"?"+url.Values{"api_key": {apiKey}}.Encode(), &r); err != nil {
		return domain.Picture{}, err
	}
	return domain.Picture{Date: r.Date, Title: strings.TrimSpace(r.Title), Explanation: domain.Truncate(r.Explanation, 2000),
		Copyright: strings.Join(strings.Fields(r.Copyright), " "), MediaType: r.MediaType, URL: r.URL, HDURL: r.HDURL}, nil
}

// errForeignImage is returned for image URLs outside NASA.
var errForeignImage = kernel.NewPublicError(kernel.ErrInvalid, "The picture is not hosted by NASA.")

// nasaImage reports whether u is an https image on a NASA host: the URL
// comes from the API response, and only NASA's own hosts are proxied.
func nasaImage(u string) bool {
	p, err := url.Parse(u)
	if err != nil || p.Scheme != "https" {
		return false
	}
	host := strings.ToLower(p.Hostname())
	return host == "nasa.gov" || strings.HasSuffix(host, ".nasa.gov")
}

// Image implements usecase.PictureProvider: only NASA images, at most 15 MB.
func (a *APOD) Image(ctx context.Context, u string) (usecase.Image, error) {
	if !nasaImage(u) {
		return usecase.Image{}, errForeignImage
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return usecase.Image{}, fmt.Errorf("build request: %w", err)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return usecase.Image{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return usecase.Image{}, fmt.Errorf("server answered %s", res.Status)
	}
	ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if !strings.HasPrefix(ct, "image/") || ct == "image/svg+xml" {
		return usecase.Image{}, fmt.Errorf("not an image: %q", ct)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, MaxImageBytes+1))
	if err != nil {
		return usecase.Image{}, fmt.Errorf("read image: %w", err)
	}
	if len(b) > MaxImageBytes {
		return usecase.Image{}, errors.New("image too large")
	}
	return usecase.Image{Bytes: b, ContentType: ct}, nil
}
