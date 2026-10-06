package subscriptions

import (
	"net/http"

	"github.com/shatranj/backend/internal/auth"
	"github.com/shatranj/backend/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, tokens *auth.TokenManager) {
	authenticated := middleware.Authenticate(func(raw string) (middleware.Principal, error) {
		claims, err := tokens.Verify(raw)
		if err != nil {
			return middleware.Principal{}, err
		}
		return middleware.Principal{UserID: claims.Subject, Email: claims.Email}, nil
	})
	mux.HandleFunc("GET /api/v1/subscriptions/plans", handler.Plans)
	mux.Handle("GET /api/v1/subscriptions/me", authenticated(http.HandlerFunc(handler.Mine)))
	mux.Handle("GET /api/v1/subscriptions/entitlements/{entitlement}", authenticated(http.HandlerFunc(handler.Entitlement)))
	mux.Handle("DELETE /api/v1/subscriptions/me", authenticated(http.HandlerFunc(handler.Cancel)))
}
