package controller

// LayoutEntryInputDTO switches a section on or off.
type LayoutEntryInputDTO struct {
	Kind    string `json:"kind"`
	Enabled bool   `json:"enabled"`
}

// SettingsInputDTO is the body of PUT /api/briefing/settings.
type SettingsInputDTO struct {
	Layout    []LayoutEntryInputDTO `json:"layout"`
	PushOn    bool                  `json:"pushOn"`
	PushAt    string                `json:"pushAt"`
	PushDays  []int                 `json:"pushDays"`
	SkipEmpty bool                  `json:"skipEmpty"`
	Lat       float64               `json:"lat"`
	Lon       float64               `json:"lon"`
}
