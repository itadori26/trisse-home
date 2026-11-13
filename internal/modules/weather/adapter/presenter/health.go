package presenter

import (
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/modules/weather/usecase"
)

// PollenRegionDTO is a selectable DWD pollen region.
type PollenRegionDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// HealthOptionsDTO is returned by GET /api/weather/health/options.
type HealthOptionsDTO struct {
	Regions []PollenRegionDTO `json:"regions"`
	Cities  []string          `json:"cities"`
}

// PollenDayDTO is the forecast of one allergen; levels use the DWD notation
// ("0", "0-1", … "3", "" = no data), burden is none, low, moderate, high or unknown.
type PollenDayDTO struct {
	Type           string `json:"type"`
	Today          string `json:"today"`
	Tomorrow       string `json:"tomorrow"`
	TodayBurden    string `json:"todayBurden"`
	TomorrowBurden string `json:"tomorrowBurden"`
}

// PollenDTO is the pollen forecast of a region.
type PollenDTO struct {
	Region    PollenRegionDTO `json:"region"`
	Days      []PollenDayDTO  `json:"days"`
	UpdatedAt *time.Time      `json:"updatedAt"`
}

// UVDTO is the UV index forecast of a city (-1 = no data); levels are
// low, moderate, high, very_high, extreme or unknown.
type UVDTO struct {
	City          string     `json:"city"`
	Today         int        `json:"today"`
	Tomorrow      int        `json:"tomorrow"`
	TodayLevel    string     `json:"todayLevel"`
	TomorrowLevel string     `json:"tomorrowLevel"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

// HealthDTO is returned by GET /api/weather/health.
type HealthDTO struct {
	Pollen *PollenDTO `json:"pollen"`
	UV     *UVDTO     `json:"uv"`
}

// HealthOptions maps the options.
func HealthOptions(o usecase.HealthOptions) HealthOptionsDTO {
	dto := HealthOptionsDTO{Regions: make([]PollenRegionDTO, 0, len(o.Regions)), Cities: o.Cities}
	for _, r := range o.Regions {
		dto.Regions = append(dto.Regions, PollenRegionDTO(r))
	}
	return dto
}

// Health maps a report.
func Health(r usecase.HealthReport) HealthDTO {
	var dto HealthDTO
	if p := r.Pollen; p != nil {
		dto.Pollen = &PollenDTO{Region: PollenRegionDTO(p.Region), UpdatedAt: optionalTime(p.UpdatedAt), Days: make([]PollenDayDTO, 0, len(p.Days))}
		for _, d := range p.Days {
			dto.Pollen.Days = append(dto.Pollen.Days, PollenDayDTO{Type: string(d.Type), Today: d.Today.String(), Tomorrow: d.Tomorrow.String(),
				TodayBurden: d.Today.Burden(), TomorrowBurden: d.Tomorrow.Burden()})
		}
	}
	if u := r.UV; u != nil {
		dto.UV = &UVDTO{City: u.City, Today: u.Today, Tomorrow: u.Tomorrow, TodayLevel: domain.UVLevel(u.Today),
			TomorrowLevel: domain.UVLevel(u.Tomorrow), UpdatedAt: optionalTime(u.UpdatedAt)}
	}
	return dto
}
