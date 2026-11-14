// Package presenter maps summaries to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import "trisse-home/internal/modules/summary/usecase"

// SettingsDTO describes the setup; configured=false hides the feature (#75).
type SettingsDTO struct {
	Configured    bool   `json:"configured"`
	URL           string `json:"url"`
	Model         string `json:"model"`
	MaxInputChars int    `json:"maxInputChars"`
}

// ResultDTO is a summary.
type ResultDTO struct {
	Summary string `json:"summary"`
	Model   string `json:"model"`
	Cached  bool   `json:"cached"`
	// Truncated is true when only the first part of the text was summarised.
	Truncated bool `json:"truncated"`
}

// ModelsDTO lists the models of an Ollama server.
type ModelsDTO struct {
	Models []string `json:"models"`
}

// Settings maps the settings.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Result maps a summary.
func Result(r usecase.Result) ResultDTO {
	return ResultDTO{Summary: r.Text, Model: r.Model, Cached: r.Cached, Truncated: r.Truncated}
}
