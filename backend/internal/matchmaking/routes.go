package matchmaking

import (
	"net/http"

	"github.com/akash6354/shatranj/backend/internal/auth"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, tokens *auth.TokenManager) {
	authenticated := middleware.Authenticate(func(raw string) (middleware.Principal, error) {
		claims, err := tokens.Verify(raw)
		if err != nil {
			return middleware.Principal{}, err
		}
		return middleware.Principal{UserID: claims.Subject, Email: claims.Email}, nil
	})
	mux.Handle("POST /api/v1/matchmaking/queue", authenticated(http.HandlerFunc(handler.Join)))
	mux.Handle("DELETE /api/v1/matchmaking/queue", authenticated(http.HandlerFunc(handler.Leave)))
	mux.Handle("GET /api/v1/matchmaking/queue", authenticated(http.HandlerFunc(handler.Status)))
}
