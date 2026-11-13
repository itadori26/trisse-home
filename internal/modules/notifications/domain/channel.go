package domain

import (
	"net/mail"
	"slices"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// ChannelKind selects the delivery adapter (#87).
type ChannelKind string

// Additional channel kinds next to the built-in ntfy settings.
const (
	ChannelGotify  ChannelKind = "gotify"
	ChannelWebhook ChannelKind = "webhook"
	ChannelEmail   ChannelKind = "email"
)

// MaxChannels bounds the number of additional channels.
const MaxChannels = 10

// DefaultWebhookTemplate is the JSON body sent when no template is set.
const DefaultWebhookTemplate = `{"title":{{title}},"message":{{body}},"severity":{{severity}},"type":{{type}}}`

// Channel is an additional delivery target with its own routing.
type Channel struct {
	ID          string
	Kind        ChannelKind
	Name        string
	Enabled     bool
	MinSeverity Severity
	// Types limits the alert types; empty means all.
	Types []string
	// URL is the Gotify server or the webhook endpoint.
	URL string
	// Secret is the Gotify app token, the webhook HMAC key or the SMTP
	// password; it is encrypted at rest and never returned by the API.
	Secret string
	// Template is the webhook JSON body with {{title}}, {{body}},
	// {{severity}} and {{type}} (JSON-encoded values).
	Template string
	// SMTP settings for e-mail.
	SMTPHost string
	SMTPPort int
	Username string
	From     string
	To       string
	// TLS is "starttls" (default), "tls" (implicit, port 465) or "none".
	TLS string
}

// targetURL validates a channel URL (#191); webhooks may carry a token in
// the query.
func targetURL(raw string, opts ...kernel.TargetOption) (string, error) {
	u, err := kernel.ParseTarget("url", raw, append([]kernel.TargetOption{kernel.WithTargetPath()}, opts...)...)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Validate checks the channel and fills defaults.
func (c *Channel) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" || len(c.Name) > 60 {
		return kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	if !c.MinSeverity.Valid() {
		c.MinSeverity = SeverityInfo
	}
	for _, t := range c.Types {
		if !slices.Contains(AlertTypes, t) {
			return kernel.NewValidationErrorf("types", "Unknown alert type {type}.", "type", t)
		}
	}
	switch c.Kind {
	case ChannelGotify:
		u, err := targetURL(c.URL)
		if err != nil {
			return err
		}
		c.URL = u
	case ChannelWebhook:
		u, err := targetURL(c.URL, kernel.WithTargetQuery())
		if err != nil {
			return err
		}
		c.URL = u
		if strings.TrimSpace(c.Template) == "" {
			c.Template = DefaultWebhookTemplate
		}
		if len(c.Template) > 4000 {
			return kernel.NewValidationError("template", "The template must be at most 4000 characters long.")
		}
	case ChannelEmail:
		if c.SMTPHost == "" || strings.ContainsAny(c.SMTPHost, " /:") {
			return kernel.NewValidationError("smtpHost", "Please enter the SMTP server name.")
		}
		if c.SMTPPort == 0 {
			c.SMTPPort = 587
		}
		if c.SMTPPort < 1 || c.SMTPPort > 65535 {
			return kernel.NewValidationError("smtpPort", "The port must be between 1 and 65535.")
		}
		if _, err := mail.ParseAddress(c.From); err != nil {
			return kernel.NewValidationError("from", "Please enter a valid sender address.")
		}
		for _, to := range strings.Split(c.To, ",") {
			if _, err := mail.ParseAddress(strings.TrimSpace(to)); err != nil {
				return kernel.NewValidationError("to", "Please enter valid recipient addresses, separated by commas.")
			}
		}
		switch c.TLS {
		case "":
			c.TLS = "starttls"
		case "starttls", "tls", "none":
		default:
			return kernel.NewValidationError("tls", "TLS must be starttls, tls or none.")
		}
	default:
		return kernel.NewValidationError("kind", "The channel type must be gotify, webhook or email.")
	}
	return nil
}

// Accepts reports whether the channel routes m.
func (c Channel) Accepts(m Message) bool {
	if !c.Enabled || m.Severity.Rank() < c.MinSeverity.Rank() {
		return false
	}
	return len(c.Types) == 0 || m.Type == "digest" || slices.Contains(c.Types, m.Type)
}
