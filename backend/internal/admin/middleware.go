package admin

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/shatranj/backend/internal/auth"
	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type RoleChecker interface {
	IsAdmin(context.Context, string) (bool, error)
}

func AdminOnly(tokens *auth.TokenManager, checker RoleChecker) func(http.Handler) http.Handler {
	authenticated := middleware.Authenticate(func(raw string) (middleware.Principal, error) {
		claims, err := tokens.Verify(raw)
		if err != nil {
			return middleware.Principal{}, err
		}
		return middleware.Principal{UserID: claims.Subject, Email: claims.Email}, nil
	})
	return func(next http.Handler) http.Handler {
		return authenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := middleware.PrincipalFromContext(r.Context())
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
				return
			}
			allowed, err := checker.IsAdmin(r.Context(), principal.UserID)
			if err != nil {
				slog.ErrorContext(r.Context(), "check admin authorization", "error", err)
				httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				return
			}
			if !allowed {
				httpapi.WriteError(w, http.StatusForbidden, "forbidden", "administrator role required")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
