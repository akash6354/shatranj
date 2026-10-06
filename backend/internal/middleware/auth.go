package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/shatranj/backend/internal/httpapi"
)

type principalKey struct{}

// Principal is the authenticated identity placed in request context.
type Principal struct {
	UserID string
	Email  string
}

// Authenticate protects a handler with a bearer access token.
func Authenticate(verify func(string) (Principal, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
				return
			}
			principal, err := verify(raw)
			if err != nil || principal.UserID == "" {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
				return
			}
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), principal)))
		})
	}
}

// PrincipalFromContext retrieves an authenticated principal.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	return principal, ok
}

func withPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}
