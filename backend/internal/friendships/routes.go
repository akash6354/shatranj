package friendships

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
	mux.Handle("GET /api/v1/friends", authenticated(http.HandlerFunc(handler.List)))
	mux.Handle("GET /api/v1/friends/requests", authenticated(http.HandlerFunc(handler.Requests)))
	mux.Handle("POST /api/v1/friends/requests", authenticated(http.HandlerFunc(handler.Request)))
	mux.Handle("POST /api/v1/friends/requests/{userID}/accept", authenticated(http.HandlerFunc(handler.Accept)))
	mux.Handle("POST /api/v1/friends/requests/{userID}/reject", authenticated(http.HandlerFunc(handler.Reject)))
	mux.Handle("POST /api/v1/friends/{userID}/block", authenticated(http.HandlerFunc(handler.Block)))
	mux.Handle("DELETE /api/v1/friends/{userID}/block", authenticated(http.HandlerFunc(handler.Unblock)))
}
