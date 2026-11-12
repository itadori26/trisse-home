package controller

// RegistryInputDTO is a registry login; an empty password keeps the stored one.
type RegistryInputDTO struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateSettingsInputDTO is the body of PUT /api/docker/updates/settings.
type UpdateSettingsInputDTO struct {
	Enabled    bool               `json:"enabled"`
	SemverTags bool               `json:"semverTags"`
	Registries []RegistryInputDTO `json:"registries"`
}
