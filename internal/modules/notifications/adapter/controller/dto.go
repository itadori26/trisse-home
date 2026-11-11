package controller

// RulesInputDTO selects what is notified.
type RulesInputDTO struct {
	Services      bool    `json:"services"`
	Containers    bool    `json:"containers"`
	Disk          bool    `json:"disk"`
	DiskThreshold float64 `json:"diskThreshold"`
	TLS           bool    `json:"tls"`
	TLSDays       int     `json:"tlsDays"`
	Backups       bool    `json:"backups"`
}

// SettingsInputDTO is the body of PUT /api/notifications/settings; an empty
// token keeps the stored one.
type SettingsInputDTO struct {
	Enabled   bool          `json:"enabled"`
	ServerURL string        `json:"serverUrl"`
	Topic     string        `json:"topic"`
	Token     string        `json:"token"`
	Rules     RulesInputDTO `json:"rules"`
}
