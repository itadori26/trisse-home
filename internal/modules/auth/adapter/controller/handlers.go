package controller

import (
	"net/http"

	"trisse-home/internal/modules/auth/adapter/presenter"
	"trisse-home/internal/modules/auth/usecase"
	"trisse-home/internal/shared/httpapi"
)

// UseCases bundles the input ports the auth controller calls.
type UseCases struct {
	Status         usecase.GetStatus
	Setup          usecase.SetupAccount
	Login          usecase.Login
	Logout         usecase.Logout
	Authenticate   usecase.Authenticate
	ChangePassword usecase.ChangePassword
}

// Handlers serves the auth endpoints.
type Handlers struct {
	uc      UseCases
	cookies CookieConfig
}

// NewHandlers returns auth handlers.
func NewHandlers(uc UseCases, cookies CookieConfig) *Handlers {
	return &Handlers{uc: uc, cookies: cookies}
}

// Status handles GET /api/auth/status.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Status.Execute(r.Context(), usecase.StatusRequest{SessionToken: h.cookies.read(r)})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Status(res))
}

// Setup handles POST /api/auth/setup.
func (h *Handlers) Setup(w http.ResponseWriter, r *http.Request) {
	var dto SetupRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Setup.Execute(r.Context(), usecase.SetupRequest{
		SetupToken: dto.SetupToken, Password: dto.Password, ClientKey: clientKey(r),
	})
	h.respondWithSession(w, r, res, err)
}

// Login handles POST /api/auth/login.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var dto LoginRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Login.Execute(r.Context(), usecase.LoginRequest{Password: dto.Password, ClientKey: clientKey(r)})
	h.respondWithSession(w, r, res, err)
}

// Logout handles POST /api/auth/logout.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Logout.Execute(r.Context(), usecase.LogoutRequest{SessionToken: h.cookies.read(r)}); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.cookies.clear(w)
	w.WriteHeader(http.StatusNoContent)
}

// ChangePassword handles PUT /api/auth/password.
func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var dto ChangePasswordRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.ChangePassword.Execute(r.Context(), usecase.ChangePasswordRequest{
		CurrentPassword: dto.CurrentPassword, NewPassword: dto.NewPassword, ClientKey: clientKey(r),
	})
	h.respondWithSession(w, r, res, err)
}

func (h *Handlers) respondWithSession(w http.ResponseWriter, r *http.Request, res usecase.SessionResponse, err error) {
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.cookies.set(w, res.SessionToken, res.ExpiresAt)
	httpapi.WriteJSON(w, http.StatusOK, presenter.Session(res))
}

func clientKey(r *http.Request) string {
	return httpapi.ClientIP(r.Context()).String()
}
