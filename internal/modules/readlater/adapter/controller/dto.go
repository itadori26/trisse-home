package controller

// SettingsInputDTO is the body of PUT /api/readlater/settings.
type SettingsInputDTO struct {
	ArchiveDays int `json:"archiveDays"`
}
