package controller

// SourceInputDTO is the body of POST/PUT /api/backup-status/sources.
type SourceInputDTO struct {
	Name        string  `json:"name"`
	MaxAgeHours float64 `json:"maxAgeHours"`
	// Kind is "backup" (default) or "check".
	Kind string `json:"kind,omitempty"`
	// Cron replaces maxAgeHours, e.g. "0 3 * * *".
	Cron         string  `json:"cron,omitempty"`
	GraceMinutes float64 `json:"graceMinutes,omitempty"`
}
