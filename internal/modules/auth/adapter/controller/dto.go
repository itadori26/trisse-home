package controller

// SetupRequestDTO is the body of POST /api/auth/setup.
type SetupRequestDTO struct {
	SetupToken string `json:"setupToken"`
	Password   string `json:"password"`
}

// LoginRequestDTO is the body of POST /api/auth/login.
type LoginRequestDTO struct {
	// Username is empty for the owner.
	Username string `json:"username,omitempty"`
	Password string `json:"password"`
	// Code is the TOTP or recovery code when two-factor authentication is enabled.
	Code string `json:"code,omitempty"`
}

// TwoFactorCodeDTO is the body of POST /api/auth/totp/confirm.
type TwoFactorCodeDTO struct {
	Code string `json:"code"`
}

// DisableTwoFactorDTO is the body of POST /api/auth/totp/disable.
type DisableTwoFactorDTO struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

// ChangePasswordRequestDTO is the body of PUT /api/auth/password.
type ChangePasswordRequestDTO struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// GuestRequestDTO creates or updates a guest; an empty password keeps it.
type GuestRequestDTO struct {
	Name     string   `json:"name,omitempty"`
	Password string   `json:"password,omitempty"`
	Tabs     []string `json:"tabs"`
	Disabled bool     `json:"disabled"`
}

// TokenRequestDTO creates an API token; ExpiresInDays 0 means never.
type TokenRequestDTO struct {
	Name          string   `json:"name"`
	Scopes        []string `json:"scopes"`
	ExpiresInDays int      `json:"expiresInDays"`
}

// SSOSecondFactorDTO finishes an SSO sign-in that needs the local code.
type SSOSecondFactorDTO struct {
	Ticket string `json:"ticket"`
	Code   string `json:"code"`
}

// PasskeyRegistrationDTO carries navigator.credentials.create() (base64url).
type PasskeyRegistrationDTO struct {
	Name              string `json:"name"`
	ClientDataJSON    string `json:"clientDataJSON"`
	AttestationObject string `json:"attestationObject"`
}

// PasskeyLoginDTO carries navigator.credentials.get() (base64url) and the
// ticket from the password step.
type PasskeyLoginDTO struct {
	Ticket            string `json:"ticket"`
	CredentialID      string `json:"credentialId"`
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
}

// KioskRequestDTO creates or changes a kiosk display (#50).
type KioskRequestDTO struct {
	Name          string   `json:"name"`
	Tabs          []string `json:"tabs"`
	Disabled      bool     `json:"disabled"`
	RotateSeconds int      `json:"rotateSeconds"`
	NightFrom     string   `json:"nightFrom"`
	NightTo       string   `json:"nightTo"`
}

// KioskLoginDTO is the body of POST /api/auth/kiosk/login.
type KioskLoginDTO struct {
	Token string `json:"token"`
}
