package controller

// SubscribeDTO is the body of POST /api/rss/feeds.
type SubscribeDTO struct {
	URL string `json:"url"`
}

// ReadDTO is the body of PUT /api/rss/items/{id}/read.
type ReadDTO struct {
	Read bool `json:"read"`
}

// MarkAllReadDTO is the body of POST /api/rss/read-all; an empty feedId marks every feed.
type MarkAllReadDTO struct {
	FeedID string `json:"feedId"`
}
