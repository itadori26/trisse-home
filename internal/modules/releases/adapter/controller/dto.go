package controller

// FollowDTO is the body of POST /api/releases/artists.
type FollowDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
