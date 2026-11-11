package controller

// SetupRequestDTO is the body of POST /api/auth/setup.
type SetupRequestDTO struct {
	SetupToken string `json:"setupToken"`
	Password   string `json:"password"`
}

// LoginRequestDTO is the body of POST /api/auth/login.
type LoginRequestDTO struct {
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
