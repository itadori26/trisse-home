package gateway

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"trisse-home/internal/modules/media/domain"
)

// tick is the unit of Jellyfin's durations.
const tick = 100 * time.Nanosecond

// Jellyfin implements usecase.MediaServerProvider with the Jellyfin REST
// API and an API key (Dashboard → API keys).
type Jellyfin struct{ client *http.Client }

// NewJellyfin returns a client.
func NewJellyfin(client *http.Client) *Jellyfin { return &Jellyfin{client: client} }

func jellyfinAuth(token string) func(*http.Request) {
	return func(r *http.Request) { r.Header.Set("Authorization", `MediaBrowser Token="`+token+`"`) }
}

type jellyfinItem struct {
	ID             string    `json:"Id"`
	Name           string    `json:"Name"`
	SeriesName     string    `json:"SeriesName"`
	AlbumArtist    string    `json:"AlbumArtist"`
	Type           string    `json:"Type"`
	ProductionYear int       `json:"ProductionYear"`
	DateCreated    time.Time `json:"DateCreated"`
	RunTimeTicks   int64     `json:"RunTimeTicks"`
}

// Overview implements usecase.MediaServerProvider.
func (j *Jellyfin) Overview(ctx context.Context, s domain.Settings) (domain.Overview, error) {
	auth := jellyfinAuth(s.Token)
	var info struct {
		ServerName string `json:"ServerName"`
	}
	if err := getJSON(ctx, j.client, s, "/System/Info", auth, &info); err != nil {
		return domain.Overview{}, err
	}
	q := url.Values{"SortBy": {"DateCreated"}, "SortOrder": {"Descending"}, "Recursive": {"true"},
		"IncludeItemTypes": {"Movie,Episode,MusicAlbum"}, "Limit": {"12"}, "Fields": {"DateCreated"}}
	var items struct {
		Items []jellyfinItem `json:"Items"`
	}
	if err := getJSON(ctx, j.client, s, "/Items?"+q.Encode(), auth, &items); err != nil {
		return domain.Overview{}, err
	}
	var sessions []struct {
		UserName       string        `json:"UserName"`
		DeviceName     string        `json:"DeviceName"`
		NowPlayingItem *jellyfinItem `json:"NowPlayingItem"`
		PlayState      struct {
			PositionTicks int64 `json:"PositionTicks"`
			IsPaused      bool  `json:"IsPaused"`
		} `json:"PlayState"`
	}
	if err := getJSON(ctx, j.client, s, "/Sessions?ActiveWithinSeconds=960", auth, &sessions); err != nil {
		return domain.Overview{}, err
	}
	o := domain.Overview{ServerName: info.ServerName}
	for _, it := range items.Items {
		o.Recent = append(o.Recent, domain.Item{ID: it.ID, Title: it.Name, Series: first(it.SeriesName, it.AlbumArtist),
			Kind: kindOf(it.Type), Year: it.ProductionYear, AddedAt: it.DateCreated})
	}
	for _, se := range sessions {
		if se.NowPlayingItem == nil {
			continue
		}
		it := se.NowPlayingItem
		o.Sessions = append(o.Sessions, domain.Session{User: se.UserName, Device: se.DeviceName, Title: it.Name,
			Series: first(it.SeriesName, it.AlbumArtist), Kind: kindOf(it.Type), Paused: se.PlayState.IsPaused,
			Position: ticks(se.PlayState.PositionTicks), Duration: ticks(it.RunTimeTicks)})
	}
	return o, nil
}

// Poster implements usecase.MediaServerProvider.
func (j *Jellyfin) Poster(ctx context.Context, s domain.Settings, itemID string) (domain.Poster, error) {
	return getImage(ctx, j.client, s.URL+"/Items/"+url.PathEscape(itemID)+"/Images/Primary?maxHeight=300&quality=80", jellyfinAuth(s.Token))
}

func ticks(t int64) time.Duration { return time.Duration(t) * tick }

// kindOf maps provider item types to the widget's kinds.
func kindOf(t string) string {
	switch t {
	case "Movie", "movie":
		return "movie"
	case "Episode", "episode":
		return "episode"
	case "Series", "show", "season", "Season":
		return "series"
	case "MusicAlbum", "album":
		return "album"
	case "Audio", "track":
		return "track"
	}
	return "other"
}

func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
