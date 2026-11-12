package presenter

import (
	"time"

	"trisse-home/internal/modules/docker/usecase"
)

// ImageUpdateDTO is one running image with its update state (#107).
type ImageUpdateDTO struct {
	Ref        string    `json:"ref"`
	Containers []string  `json:"containers"`
	Status     string    `json:"status" tstype:"'unknown' | 'up_to_date' | 'update_available' | 'pinned' | 'local' | 'error'"`
	NewerTag   string    `json:"newerTag"`
	ReleaseURL string    `json:"releaseUrl"`
	CheckedAt  time.Time `json:"checkedAt"`
	Error      string    `json:"error"`
}

// ImageUpdates maps the update list.
func ImageUpdates(list []usecase.ImageUpdateView) []ImageUpdateDTO {
	out := make([]ImageUpdateDTO, 0, len(list))
	for _, v := range list {
		out = append(out, ImageUpdateDTO{Ref: v.Ref, Containers: v.Containers, Status: string(v.Status), NewerTag: v.NewerTag,
			ReleaseURL: v.ReleaseURL, CheckedAt: v.CheckedAt, Error: v.Error})
	}
	return out
}

// RegistryDTO is a stored registry login; the password is never returned.
type RegistryDTO struct {
	Host        string `json:"host"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
}

// UpdateSettingsDTO is returned by GET/PUT /api/docker/updates/settings.
type UpdateSettingsDTO struct {
	Enabled    bool          `json:"enabled"`
	SemverTags bool          `json:"semverTags"`
	Registries []RegistryDTO `json:"registries"`
}

// UpdateSettings maps the settings view.
func UpdateSettings(v usecase.UpdateSettingsView) UpdateSettingsDTO {
	dto := UpdateSettingsDTO{Enabled: v.Enabled, SemverTags: v.SemverTags, Registries: make([]RegistryDTO, 0, len(v.Registries))}
	for _, r := range v.Registries {
		dto.Registries = append(dto.Registries, RegistryDTO(r))
	}
	return dto
}
