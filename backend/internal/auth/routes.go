package auth

import (
	"net/http"
	"time"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, tokens *TokenManager) {
	limiter := middleware.RateLimit(10, time.Minute)
	mux.Handle("POST /api/v1/auth/register", limiter(http.HandlerFunc(handler.Register)))
	mux.Handle("POST /api/v1/auth/login", limiter(http.HandlerFunc(handler.Login)))
	authenticated := middleware.Authenticate(func(raw string) (middleware.Principal, error) {
		claims, err := tokens.Verify(raw)
		if err != nil {
			return middleware.Principal{}, err
		}
		return middleware.Principal{UserID: claims.Subject, Email: claims.Email}, nil
	})
	mux.Handle("GET /api/v1/auth/validate", authenticated(http.HandlerFunc(handler.Validate)))
	// The OTP delivery contract is prepared in otp.go; delivery and verification
	// stay unavailable until an email provider and durable OTP store are wired.
	mux.Handle("POST /api/v1/auth/otp/request", limiter(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpapi.WriteError(w, http.StatusNotImplemented, "otp_unavailable", "email code login is not configured")
	})))
}
