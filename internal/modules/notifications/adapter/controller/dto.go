package controller

import "time"

// RulesInputDTO selects what is notified.
type RulesInputDTO struct {
	Services      bool    `json:"services"`
	Containers    bool    `json:"containers"`
	Disk          bool    `json:"disk"`
	DiskThreshold float64 `json:"diskThreshold"`
	TLS           bool    `json:"tls"`
	TLSDays       int     `json:"tlsDays"`
	Backups       bool    `json:"backups"`
	Security      bool    `json:"security"`
}

// SettingsInputDTO is the body of PUT /api/notifications/settings; an empty
// token keeps the stored one.
type SettingsInputDTO struct {
	Enabled   bool          `json:"enabled"`
	ServerURL string        `json:"serverUrl"`
	Topic     string        `json:"topic"`
	Token     string        `json:"token"`
	Rules     RulesInputDTO `json:"rules"`
	Policy    PolicyDTO     `json:"policy"`
}

// PolicyDTO configures severities per alert type and quiet hours (#69).
type PolicyDTO struct {
	Severities map[string]string `json:"severities"`
	Quiet      QuietHoursDTO     `json:"quiet"`
}

// QuietHoursDTO mutes non-critical notifications; days are 0 (Sunday) to 6.
type QuietHoursDTO struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Days    []int  `json:"days"`
}

// ChannelDTO creates or updates a channel (#87); an empty secret keeps the stored one.
type ChannelDTO struct {
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	MinSeverity string   `json:"minSeverity"`
	Types       []string `json:"types"`
	URL         string   `json:"url"`
	Secret      string   `json:"secret"`
	Template    string   `json:"template"`
	SMTPHost    string   `json:"smtpHost"`
	SMTPPort    int      `json:"smtpPort"`
	Username    string   `json:"username"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	TLS         string   `json:"tls"`
}

// MaintenanceDTO starts a maintenance window; without start it starts now
// and lasts Minutes.
type MaintenanceDTO struct {
	Scope   []string   `json:"scope"`
	Start   *time.Time `json:"start"`
	End     *time.Time `json:"end"`
	Minutes int        `json:"minutes"`
	Note    string     `json:"note"`
}
