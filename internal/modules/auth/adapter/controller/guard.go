package controller

import (
	"crypto/subtle"
	"errors"
	"net/http"

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
		ctx := usecase.WithClient(r.Context(), domain.NewClient(r.UserAgent(), clientKey(r)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func validCSRF(got, want string) bool {
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
