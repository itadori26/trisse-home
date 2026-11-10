package controller

// SetupRequestDTO is the body of POST /api/auth/setup.
type SetupRequestDTO struct {
	SetupToken string `json:"setupToken"`
	Password   string `json:"password"`
}

// LoginRequestDTO is the body of POST /api/auth/login.
type LoginRequestDTO struct {
	Password string `json:"password"`
}

// ChangePasswordRequestDTO is the body of PUT /api/auth/password.
type ChangePasswordRequestDTO struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}
