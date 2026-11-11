package controller

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"

	"trisse-home/internal/modules/auth/adapter/presenter"
	"trisse-home/internal/modules/auth/domain"
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
	Sessions       usecase.ManageSessions
	TwoFactor      usecase.TwoFactor
	SecurityLog    usecase.SecurityLog
	Authorizer     usecase.Authorizer
	Guests         usecase.Guests
	APITokens      usecase.APITokens
	// SSO is nil when no single sign-on method is configured.
	SSO usecase.SSO
	// DemoLogin is nil unless the instance runs in demo mode.
	DemoLogin usecase.DemoLogin
}

// Handlers serves the auth endpoints.
type Handlers struct {
	uc      UseCases
	cookies CookieConfig
	proxy   ProxyAuth
}

// ProxyAuth configures login via a header set by a trusted reverse proxy.
type ProxyAuth struct {
	// Header carries the authenticated user name, e.g. "Remote-User".
	Header string
	// Proxies are the only peers whose header is believed.
	Proxies []netip.Prefix
}

// WithProxyAuth enables the trusted header login.
func (h *Handlers) WithProxyAuth(p ProxyAuth) *Handlers { h.proxy = p; return h }

// fromTrustedProxy reports whether the TCP peer (not a forwarded address)
// is one of the configured proxies. Headers from anyone else are ignored.
func (h *Handlers) fromTrustedProxy(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	addr = addr.Unmap()
	for _, p := range h.proxy.Proxies {
		if p.Contains(addr) {
			return true
		}
	}
	return false
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
	dto := presenter.Status(res)
	dto.Demo = h.uc.DemoLogin != nil
	if h.uc.SSO != nil {
		st := h.uc.SSO.Status()
		dto.SSO = presenter.SSOStatusDTO{OIDC: st.OIDC, Header: st.Header && h.proxy.Header != "" && h.fromTrustedProxy(r), LocalLogin: st.LocalLogin}
	} else {
		dto.SSO = presenter.SSOStatusDTO{LocalLogin: true}
	}
	httpapi.WriteJSON(w, http.StatusOK, dto)
}

// DemoLogin handles POST /api/auth/demo (only routed in demo mode).
func (h *Handlers) DemoLogin(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.DemoLogin.Execute(r.Context(), usecase.DemoLoginRequest{ClientKey: clientKey(r), UserAgent: r.UserAgent()})
	h.respondWithSession(w, r, res, err)
}

// Setup handles POST /api/auth/setup.
func (h *Handlers) Setup(w http.ResponseWriter, r *http.Request) {
	var dto SetupRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Setup.Execute(r.Context(), usecase.SetupRequest{
		SetupToken: dto.SetupToken, Password: dto.Password, ClientKey: clientKey(r), UserAgent: r.UserAgent(),
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
	res, err := h.uc.Login.Execute(r.Context(), usecase.LoginRequest{Username: dto.Username, Password: dto.Password, TOTPCode: dto.Code, ClientKey: clientKey(r), UserAgent: r.UserAgent()})
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
		CurrentPassword: dto.CurrentPassword, NewPassword: dto.NewPassword, ClientKey: clientKey(r), UserAgent: r.UserAgent(),
	})
	h.respondWithSession(w, r, res, err)
}

// ListSessions handles GET /api/auth/sessions.
func (h *Handlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Sessions.List(r.Context(), h.cookies.read(r))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Sessions(res))
}

// RevokeSession handles DELETE /api/auth/sessions/{id}.
func (h *Handlers) RevokeSession(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Sessions.Revoke(r.Context(), h.cookies.read(r), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RevokeOtherSessions handles POST /api/auth/sessions/revoke-others.
func (h *Handlers) RevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.Sessions.RevokeOthers(r.Context(), h.cookies.read(r))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.RevokedDTO{Revoked: n})
}

// TwoFactorStatus handles GET /api/auth/totp.
func (h *Handlers) TwoFactorStatus(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.TwoFactor.Status(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.TwoFactorStatusDTO{Enabled: res.Enabled, RecoveryCodesLeft: res.RecoveryCodesLeft})
}

// StartTwoFactor handles POST /api/auth/totp/setup.
func (h *Handlers) StartTwoFactor(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.TwoFactor.Start(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, presenter.EnrollmentDTO{Secret: res.Secret, URI: res.URI})
}

// ConfirmTwoFactor handles POST /api/auth/totp/confirm.
func (h *Handlers) ConfirmTwoFactor(w http.ResponseWriter, r *http.Request) {
	var dto TwoFactorCodeDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	codes, err := h.uc.TwoFactor.Confirm(r.Context(), dto.Code)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, presenter.RecoveryCodesDTO{Codes: codes})
}

// DisableTwoFactor handles POST /api/auth/totp/disable.
func (h *Handlers) DisableTwoFactor(w http.ResponseWriter, r *http.Request) {
	var dto DisableTwoFactorDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	err := h.uc.TwoFactor.Disable(r.Context(), usecase.DisableTwoFactorRequest{Password: dto.Password, Code: dto.Code, ClientKey: clientKey(r)})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

// SecurityEvents handles GET /api/auth/events.
func (h *Handlers) SecurityEvents(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.SecurityLog.List(r.Context(), 200)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SecurityEvents(list))
}

// ListGuests handles GET /api/auth/guests.
func (h *Handlers) ListGuests(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.Guests.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Guests(list))
}

// CreateGuest handles POST /api/auth/guests.
func (h *Handlers) CreateGuest(w http.ResponseWriter, r *http.Request) {
	var dto GuestRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	g, err := h.uc.Guests.Create(r.Context(), usecase.NewGuestRequest{Name: dto.Name, Password: dto.Password, Tabs: dto.Tabs})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Guest(g))
}

// UpdateGuest handles PUT /api/auth/guests/{id}.
func (h *Handlers) UpdateGuest(w http.ResponseWriter, r *http.Request) {
	var dto GuestRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	g, err := h.uc.Guests.Update(r.Context(), usecase.UpdateGuestRequest{ID: r.PathValue("id"), Tabs: dto.Tabs, Disabled: dto.Disabled, Password: dto.Password})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Guest(g))
}

// DeleteGuest handles DELETE /api/auth/guests/{id}.
func (h *Handlers) DeleteGuest(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Guests.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListTokens handles GET /api/auth/tokens.
func (h *Handlers) ListTokens(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.APITokens.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Tokens(list))
}

// CreateToken handles POST /api/auth/tokens; the token is returned once.
func (h *Handlers) CreateToken(w http.ResponseWriter, r *http.Request) {
	var dto TokenRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	t, err := h.uc.APITokens.Create(r.Context(), usecase.NewTokenRequest{Name: dto.Name, Scopes: dto.Scopes, ExpiresInDays: dto.ExpiresInDays})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.CreatedToken(t))
}

// RevokeToken handles DELETE /api/auth/tokens/{id}.
func (h *Handlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.APITokens.Revoke(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

const oidcStateCookie = "trisse_oidc_state"

// OIDCStart handles GET /api/auth/oidc/start: redirect to the provider and
// bind the state to this browser (protects against login CSRF).
func (h *Handlers) OIDCStart(w http.ResponseWriter, r *http.Request) {
	if h.uc.SSO == nil {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	url, state, err := h.uc.SSO.BeginOIDC(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	// SameSite=Lax: the provider redirects back with a top-level GET.
	// Secure follows COOKIE_SECURE like the session cookie (plain HTTP in the LAN by default).
	http.SetCookie(w, &http.Cookie{Name: oidcStateCookie, Value: state, Path: "/api/auth/oidc/", HttpOnly: true, //nolint:gosec // see above
		Secure: h.cookies.Secure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	http.Redirect(w, r, url, http.StatusFound)
}

// OIDCCallback handles GET /api/auth/oidc/callback.
func (h *Handlers) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	clearState := &http.Cookie{Name: oidcStateCookie, Path: "/api/auth/oidc/", MaxAge: -1, HttpOnly: true, //nolint:gosec // Secure follows COOKIE_SECURE
		Secure: h.cookies.Secure, SameSite: http.SameSiteLaxMode}
	state := r.URL.Query().Get("state")
	c, err := r.Cookie(oidcStateCookie)
	if h.uc.SSO == nil || err != nil || state == "" || !validCSRF(c.Value, state) {
		http.SetCookie(w, clearState)
		http.Redirect(w, r, "/?sso_error=state", http.StatusFound)
		return
	}
	http.SetCookie(w, clearState)
	if e := r.URL.Query().Get("error"); e != "" {
		http.Redirect(w, r, "/?sso_error=provider", http.StatusFound)
		return
	}
	client := domain.NewClient(r.UserAgent(), clientKey(r))
	res, ticket, err := h.uc.SSO.CompleteOIDC(r.Context(), state, r.URL.Query().Get("code"), client)
	h.redirectAfterSSO(w, r, res, ticket, err)
}

func (h *Handlers) redirectAfterSSO(w http.ResponseWriter, r *http.Request, res usecase.SessionResponse, ticket string, err error) {
	switch {
	case errors.Is(err, domain.ErrSSOSecondFactorNeed):
		http.Redirect(w, r, "/?sso_ticket="+url.QueryEscape(ticket), http.StatusFound)
	case errors.Is(err, domain.ErrSSONotAllowed):
		http.Redirect(w, r, "/?sso_error=not_allowed", http.StatusFound)
	case err != nil:
		http.Redirect(w, r, "/?sso_error=failed", http.StatusFound)
	default:
		h.cookies.set(w, res.SessionToken, res.ExpiresAt)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// HeaderLogin handles POST /api/auth/sso/header: the trusted proxy already
// authenticated the user and passes the name in the configured header.
func (h *Handlers) HeaderLogin(w http.ResponseWriter, r *http.Request) {
	if h.uc.SSO == nil || h.proxy.Header == "" || !h.fromTrustedProxy(r) {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	client := domain.NewClient(r.UserAgent(), clientKey(r))
	res, ticket, err := h.uc.SSO.HeaderLogin(r.Context(), r.Header.Get(h.proxy.Header), client)
	if errors.Is(err, domain.ErrSSOSecondFactorNeed) {
		httpapi.WriteJSON(w, http.StatusAccepted, presenter.SSOTicketDTO{Ticket: ticket})
		return
	}
	h.respondWithSession(w, r, res, err)
}

// SSOSecondFactor handles POST /api/auth/sso/2fa.
func (h *Handlers) SSOSecondFactor(w http.ResponseWriter, r *http.Request) {
	var dto SSOSecondFactorDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if h.uc.SSO == nil {
		httpapi.WriteDomainError(w, r, domain.ErrSSOUnavailable)
		return
	}
	res, err := h.uc.SSO.CompleteSecondFactor(r.Context(), dto.Ticket, dto.Code, domain.NewClient(r.UserAgent(), clientKey(r)))
	h.respondWithSession(w, r, res, err)
}
