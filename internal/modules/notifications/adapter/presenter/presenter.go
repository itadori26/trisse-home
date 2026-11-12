// Package presenter maps notification settings to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/modules/notifications/usecase"
)

// RulesDTO selects what is notified.
type RulesDTO struct {
	Services      bool    `json:"services"`
	Containers    bool    `json:"containers"`
	Disk          bool    `json:"disk"`
	DiskThreshold float64 `json:"diskThreshold"`
	TLS           bool    `json:"tls"`
	TLSDays       int     `json:"tlsDays"`
	Backups       bool    `json:"backups"`
	Security      bool    `json:"security"`
}

// SettingsDTO is returned by the settings endpoints; the token is never returned.
type SettingsDTO struct {
	Enabled       bool       `json:"enabled"`
	ServerURL     string     `json:"serverUrl"`
	Topic         string     `json:"topic"`
	HasToken      bool       `json:"hasToken"`
	Rules         RulesDTO   `json:"rules"`
	Policy        PolicyDTO  `json:"policy"`
	LastError     string     `json:"lastError"`
	LastDelivered *time.Time `json:"lastDelivered"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO {
	dto := SettingsDTO{Enabled: v.Enabled, ServerURL: v.ServerURL, Topic: v.Topic, HasToken: v.HasToken,
		Rules: RulesDTO(v.Rules), Policy: policy(v.Policy), LastError: v.LastError}
	if !v.LastDelivered.IsZero() {
		t := v.LastDelivered
		dto.LastDelivered = &t
	}
	return dto
}

// PolicyDTO holds severities per alert type (all types, defaults filled in)
// and quiet hours.
type PolicyDTO struct {
	Severities map[string]string `json:"severities"`
	Quiet      QuietHoursDTO     `json:"quiet"`
}

// QuietHoursDTO holds the quiet hours; days are 0 (Sunday) to 6.
type QuietHoursDTO struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Days    []int  `json:"days"`
}

func policy(p domain.Policy) PolicyDTO {
	dto := PolicyDTO{Severities: map[string]string{}, Quiet: QuietHoursDTO{Enabled: p.Quiet.Enabled, Start: p.Quiet.Start, End: p.Quiet.End, Days: []int{}}}
	for _, t := range domain.AlertTypes {
		dto.Severities[t] = string(p.SeverityOf(t))
	}
	for _, d := range p.Quiet.Days {
		dto.Quiet.Days = append(dto.Quiet.Days, int(d))
	}
	return dto
}

// ChannelViewDTO describes a channel; the secret is never returned.
type ChannelViewDTO struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	MinSeverity string   `json:"minSeverity"`
	Types       []string `json:"types"`
	URL         string   `json:"url"`
	HasSecret   bool     `json:"hasSecret"`
	Template    string   `json:"template"`
	SMTPHost    string   `json:"smtpHost"`
	SMTPPort    int      `json:"smtpPort"`
	Username    string   `json:"username"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	TLS         string   `json:"tls"`
}

// Channel maps a channel.
func Channel(c usecase.ChannelView) ChannelViewDTO {
	types := c.Types
	if types == nil {
		types = []string{}
	}
	return ChannelViewDTO{ID: c.ID, Kind: string(c.Kind), Name: c.Name, Enabled: c.Enabled, MinSeverity: string(c.MinSeverity),
		Types: types, URL: c.URL, HasSecret: c.HasSecret, Template: c.Template, SMTPHost: c.SMTPHost, SMTPPort: c.SMTPPort,
		Username: c.Username, From: c.From, To: c.To, TLS: c.TLS}
}

// Channels maps the channel list.
func Channels(list []usecase.ChannelView) []ChannelViewDTO {
	out := make([]ChannelViewDTO, 0, len(list))
	for _, c := range list {
		out = append(out, Channel(c))
	}
	return out
}

// MaintenanceWindowDTO describes a maintenance window.
type MaintenanceWindowDTO struct {
	ID    string    `json:"id"`
	Scope []string  `json:"scope"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Note  string    `json:"note"`
}

// Maintenance maps maintenance windows.
func Maintenance(list []domain.MaintenanceWindow) []MaintenanceWindowDTO {
	out := make([]MaintenanceWindowDTO, 0, len(list))
	for _, w := range list {
		out = append(out, MaintenanceWindowDTO{ID: w.ID, Scope: w.Scope, Start: w.Start, End: w.End, Note: w.Note})
	}
	return out
}
