package controller

// RuleInputDTO maps summaries containing match to a label, color and icon.
type RuleInputDTO struct {
	Match string `json:"match"`
	Label string `json:"label"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// SettingsInputDTO is the body of PUT /api/waste/settings.
type SettingsInputDTO struct {
	URL   string         `json:"url"`
	Rules []RuleInputDTO `json:"rules"`
}
