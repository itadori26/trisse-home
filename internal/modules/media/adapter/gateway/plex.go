package gateway

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"trisse-home/internal/modules/media/domain"
)

// Plex implements usecase.MediaServerProvider with the Plex Media Server
// API and an X-Plex-Token.
type Plex struct{ client *http.Client }

// NewPlex returns a client.
func NewPlex(client *http.Client) *Plex { return &Plex{client: client} }

func plexAuth(token string) func(*http.Request) {
	return func(r *http.Request) { r.Header.Set("X-Plex-Token", token) }
}

type plexMetadata struct {
	RatingKey        string `json:"ratingKey"`
	Title            string `json:"title"`
	GrandparentTitle string `json:"grandparentTitle"`
	ParentTitle      string `json:"parentTitle"`
	Type             string `json:"type"`
	Year             int    `json:"year"`
	AddedAt          int64  `json:"addedAt"`
	Duration         int64  `json:"duration"`
	ViewOffset       int64  `json:"viewOffset"`
	User             struct {
		Title string `json:"title"`
	} `json:"User"`
	Player struct {
		Title string `json:"title"`
		State string `json:"state"`
	} `json:"Player"`
}

type plexContainer struct {
	MediaContainer struct {
		FriendlyName string         `json:"friendlyName"`
		Metadata     []plexMetadata `json:"Metadata"`
	} `json:"MediaContainer"`
}

// Overview implements usecase.MediaServerProvider.
func (p *Plex) Overview(ctx context.Context, s domain.Settings) (domain.Overview, error) {
	auth := plexAuth(s.Token)
	var root, recent, sessions plexContainer
	if err := getJSON(ctx, p.client, s, "/", auth, &root); err != nil {
		return domain.Overview{}, err
	}
	if err := getJSON(ctx, p.client, s, "/library/recentlyAdded?X-Plex-Container-Start=0&X-Plex-Container-Size=12", auth, &recent); err != nil {
		return domain.Overview{}, err
	}
	if err := getJSON(ctx, p.client, s, "/status/sessions", auth, &sessions); err != nil {
		return domain.Overview{}, err
	}
	o := domain.Overview{ServerName: root.MediaContainer.FriendlyName}
	for _, m := range recent.MediaContainer.Metadata {
		o.Recent = append(o.Recent, domain.Item{ID: m.RatingKey, Title: m.Title, Series: first(m.GrandparentTitle, m.ParentTitle),
			Kind: kindOf(m.Type), Year: m.Year, AddedAt: time.Unix(m.AddedAt, 0).UTC()})
	}
	for _, m := range sessions.MediaContainer.Metadata {
		o.Sessions = append(o.Sessions, domain.Session{User: m.User.Title, Device: m.Player.Title, Title: m.Title,
			Series: first(m.GrandparentTitle, m.ParentTitle), Kind: kindOf(m.Type), Paused: m.Player.State == "paused",
			Position: time.Duration(m.ViewOffset) * time.Millisecond, Duration: time.Duration(m.Duration) * time.Millisecond})
	}
	return o, nil
}

// Poster implements usecase.MediaServerProvider via the photo transcoder,
// which returns a small JPEG instead of the full-size artwork.
func (p *Plex) Poster(ctx context.Context, s domain.Settings, itemID string) (domain.Poster, error) {
	q := url.Values{"width": {"200"}, "height": {"300"}, "minSize": {"1"}, "url": {"/library/metadata/" + itemID + "/thumb"}}
	return getImage(ctx, p.client, s.URL+"/photo/:/transcode?"+q.Encode(), plexAuth(s.Token))
}
