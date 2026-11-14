package controller

// SettingsInputDTO is the body of PUT /api/summary/settings.
type SettingsInputDTO struct {
	URL           string `json:"url"`
	Model         string `json:"model"`
	MaxInputChars int    `json:"maxInputChars"`
}

// ModelsInputDTO is the body of POST /api/summary/models.
type ModelsInputDTO struct {
	URL string `json:"url"`
}

// SummarizeDTO is the body of POST /api/summary: a text (notes) or the link
// of a feed item whose page is loaded on the server.
type SummarizeDTO struct {
	Text     string `json:"text,omitempty"`
	URL      string `json:"url,omitempty"`
	Language string `json:"language"`
}

// AskDTO is the body of POST /api/summary/ask (#162).
type AskDTO struct {
	Question string `json:"question"`
	// Context is the compact status text the frontend built.
	Context  string `json:"context"`
	Language string `json:"language"`
}
