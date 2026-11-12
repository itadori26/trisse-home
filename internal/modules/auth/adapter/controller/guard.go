package controller

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/modules/auth/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// CSRFHeader carries the per-session CSRF token on state-changing requests.
const CSRFHeader = "X-CSRF-Token"

// Guard is middleware that only lets requests with a valid session through.
// State-changing methods additionally need the session's CSRF token, which a
// cross-site attacker cannot read (synchronizer token pattern).
func (h *Handlers) Guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw, ok := bearerToken(r); ok && h.uc.APITokens != nil {
			h.serveWithToken(w, r, raw, next)
			return
		}
		res, err := h.uc.Authenticate.Execute(r.Context(), usecase.AuthenticateRequest{SessionToken: h.cookies.read(r)})
		if errors.Is(err, domain.ErrSessionInvalid) {
			h.cookies.clear(w)
		}
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		if !isSafeMethod(r.Method) && !validCSRF(r.Header.Get(CSRFHeader), res.CSRFToken) {
			httpapi.WriteDomainError(w, r, kernel.NewPublicError(kernel.ErrForbidden, "Missing or invalid CSRF token."))
			return
		}
		if res.Refreshed {
			h.cookies.set(w, h.cookies.read(r), res.ExpiresAt)
		}
		if err := h.uc.Authorizer.Authorize(r.Context(), res.Principal, r.Method, r.URL.Path); err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		ctx := kernel.WithPrincipal(r.Context(), res.Principal)
		ctx = usecase.WithClient(ctx, domain.NewClient(r.UserAgent(), clientKey(r)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func validCSRF(got, want string) bool {
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// bearerToken returns the API token of an "Authorization: Bearer" header.
func bearerToken(r *http.Request) (string, bool) {
	scheme, raw, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(raw) == "" {
		return "", false
	}
	return strings.TrimSpace(raw), true
}

// serveWithToken authenticates an API token. No cookie is involved, so a
// browser cannot be tricked into sending it: CSRF protection is not needed.
func (h *Handlers) serveWithToken(w http.ResponseWriter, r *http.Request, raw string, next http.Handler) {
	client := domain.NewClient(r.UserAgent(), clientKey(r))
	p, err := h.uc.APITokens.Authenticate(r.Context(), raw, client)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Authorizer.Authorize(r.Context(), p, r.Method, r.URL.Path); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	ctx := usecase.WithClient(kernel.WithPrincipal(r.Context(), p), client)
	next.ServeHTTP(w, r.WithContext(ctx))
}
