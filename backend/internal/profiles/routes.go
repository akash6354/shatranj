package profiles

import (
	"net/http"

	"github.com/akash6354/shatranj/backend/internal/auth"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, repository Repository, tokens *auth.TokenManager) {
	handler := NewHandler(NewService(repository))
	authenticated := middleware.Authenticate(func(raw string) (middleware.Principal, error) {
		claims, err := tokens.Verify(raw)
		if err != nil {
			return middleware.Principal{}, err
		}
		return middleware.Principal{UserID: claims.Subject, Email: claims.Email}, nil
	})
	mux.HandleFunc("GET /api/v1/profiles/{username}", handler.Get)
	mux.Handle("GET /api/v1/profiles/me", authenticated(http.HandlerFunc(handler.Me)))
	mux.Handle("PATCH /api/v1/profiles/me", authenticated(http.HandlerFunc(handler.Update)))
}
