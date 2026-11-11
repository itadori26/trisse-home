package presenter

import (
	"time"

	"trisse-home/internal/modules/auth/usecase"
)

// StatusDTO is returned by GET /api/auth/status.
type StatusDTO struct {
	SetupRequired bool   `json:"setupRequired"`
	Authenticated bool   `json:"authenticated"`
	CSRFToken     string `json:"csrfToken,omitempty"`
	// Demo is true on a public demo instance (POST /api/auth/demo logs in).
	Demo bool `json:"demo"`
	// Role ("owner" or "guest") and Username of the signed-in account.
	Role     string `json:"role,omitempty"`
	Username string `json:"username,omitempty"`
	// SSO lists the available sign-in methods.
	SSO SSOStatusDTO `json:"sso"`
}

// SSOStatusDTO tells the login page which methods exist.
type SSOStatusDTO struct {
	OIDC       bool `json:"oidc"`
	Header     bool `json:"header"`
	LocalLogin bool `json:"localLogin"`
}

// SSOTicketDTO is returned when an SSO sign-in still needs the local code.
type SSOTicketDTO struct {
	Ticket string `json:"ticket"`
}

// SessionDTO is returned after setup, login and password change. The session
// token itself is only sent as an HttpOnly cookie.
type SessionDTO struct {
	CSRFToken string    `json:"csrfToken"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Status maps a StatusResponse to its DTO.
func Status(r usecase.StatusResponse) StatusDTO {
	return StatusDTO{SetupRequired: r.SetupRequired, Authenticated: r.Authenticated, CSRFToken: r.CSRFToken, Role: r.Role, Username: r.Username}
}

// Session maps a SessionResponse to its DTO.
func Session(r usecase.SessionResponse) SessionDTO {
	return SessionDTO{CSRFToken: r.CSRFToken, ExpiresAt: r.ExpiresAt}
}

// SessionInfoDTO describes an active session in the session overview.
type SessionInfoDTO struct {
	ID         string    `json:"id"`
	UserAgent  string    `json:"userAgent"`
	ClientIP   string    `json:"clientIp"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Current    bool      `json:"current"`
}

// RevokedDTO is returned by POST /api/auth/sessions/revoke-others.
type RevokedDTO struct {
	Revoked int64 `json:"revoked"`
}

// Sessions maps session views.
func Sessions(vs []usecase.SessionView) []SessionInfoDTO {
	out := make([]SessionInfoDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, SessionInfoDTO(v))
	}
	return out
}

// TwoFactorStatusDTO is returned by GET /api/auth/totp.
type TwoFactorStatusDTO struct {
	Enabled           bool `json:"enabled"`
	RecoveryCodesLeft int  `json:"recoveryCodesLeft"`
}

// EnrollmentDTO is returned by POST /api/auth/totp/setup; uri is encoded as QR code.
type EnrollmentDTO struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

// RecoveryCodesDTO is returned once by POST /api/auth/totp/confirm.
type RecoveryCodesDTO struct {
	Codes []string `json:"codes"`
}

// SecurityEventDTO is one entry of GET /api/auth/events.
type SecurityEventDTO struct {
	ID        string    `json:"id"`
	At        time.Time `json:"at"`
	Type      string    `json:"type"`
	AccountID string    `json:"accountId"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	Detail    string    `json:"detail"`
}

// SecurityEvents maps the security log.
func SecurityEvents(list []usecase.SecurityEventView) []SecurityEventDTO {
	out := make([]SecurityEventDTO, 0, len(list))
	for _, e := range list {
		out = append(out, SecurityEventDTO(e))
	}
	return out
}

// GuestDTO is a guest account in GET /api/auth/guests.
type GuestDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tabs      []string  `json:"tabs"`
	Disabled  bool      `json:"disabled"`
	CreatedAt time.Time `json:"createdAt"`
}

// Guest maps one guest.
func Guest(g usecase.GuestView) GuestDTO { return GuestDTO(g) }

// Guests maps the guest list.
func Guests(list []usecase.GuestView) []GuestDTO {
	out := make([]GuestDTO, 0, len(list))
	for _, g := range list {
		out = append(out, Guest(g))
	}
	return out
}

// TokenDTO describes an API token without its secret.
type TokenDTO struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	LastUsedIP string     `json:"lastUsedIp"`
}

// CreatedTokenDTO is returned once after creation.
type CreatedTokenDTO struct {
	TokenDTO
	Token string `json:"token"`
}

func optional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Token maps one token.
func Token(t usecase.TokenView) TokenDTO {
	return TokenDTO{ID: t.ID, Name: t.Name, Scopes: t.Scopes, CreatedAt: t.CreatedAt, ExpiresAt: optional(t.ExpiresAt),
		LastUsedAt: optional(t.LastUsedAt), LastUsedIP: t.LastUsedIP}
}

// Tokens maps the token list.
func Tokens(list []usecase.TokenView) []TokenDTO {
	out := make([]TokenDTO, 0, len(list))
	for _, t := range list {
		out = append(out, Token(t))
	}
	return out
}

// CreatedToken maps a new token including its secret.
func CreatedToken(t usecase.CreatedToken) CreatedTokenDTO {
	return CreatedTokenDTO{TokenDTO: Token(t.TokenView), Token: t.Token}
}
