package lessons

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
	mux.Handle("GET /api/v1/lessons", authenticated(http.HandlerFunc(handler.List)))
	mux.Handle("GET /api/v1/lessons/{courseID}", authenticated(http.HandlerFunc(handler.Get)))
	mux.Handle("PUT /api/v1/lessons/chapters/{chapterID}/progress", authenticated(http.HandlerFunc(handler.UpdateProgress)))
}
