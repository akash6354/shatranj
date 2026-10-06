package clubs

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
	mux.Handle("GET /api/v1/clubs", authenticated(http.HandlerFunc(handler.List)))
	mux.Handle("POST /api/v1/clubs", authenticated(http.HandlerFunc(handler.Create)))
	mux.Handle("GET /api/v1/clubs/{clubID}", authenticated(http.HandlerFunc(handler.Get)))
	mux.Handle("POST /api/v1/clubs/{clubID}/join", authenticated(http.HandlerFunc(handler.Join)))
	mux.Handle("DELETE /api/v1/clubs/{clubID}/membership", authenticated(http.HandlerFunc(handler.Leave)))
	mux.Handle("GET /api/v1/clubs/{clubID}/members", authenticated(http.HandlerFunc(handler.Members)))
	mux.Handle("GET /api/v1/clubs/{clubID}/requests", authenticated(http.HandlerFunc(handler.Requests)))
	mux.Handle("POST /api/v1/clubs/{clubID}/requests/{requestID}", authenticated(http.HandlerFunc(handler.ResolveRequest)))
	mux.Handle("PUT /api/v1/clubs/{clubID}/members/{userID}/role", authenticated(http.HandlerFunc(handler.SetRole)))
}
