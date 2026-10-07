package achievements

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
	mux.Handle("GET /api/v1/achievements", authenticated(http.HandlerFunc(handler.Catalog)))
	mux.Handle("GET /api/v1/achievements/me", authenticated(http.HandlerFunc(handler.Mine)))
}
