package controller

// InstanceInputDTO is the body of POST /api/arr/instances and
// PUT /api/arr/instances/{id}; an empty API key keeps the stored one.
type InstanceInputDTO struct {
	App    string `json:"app" tstype:"'sonarr' | 'radarr' | 'prowlarr'"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	APIKey string `json:"apiKey"`
}
