package puzzles

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
	mux.Handle("GET /api/v1/puzzles", authenticated(http.HandlerFunc(handler.List)))
	mux.Handle("GET /api/v1/puzzles/daily", authenticated(http.HandlerFunc(handler.Daily)))
	mux.Handle("GET /api/v1/puzzles/rating", authenticated(http.HandlerFunc(handler.Rating)))
	mux.Handle("POST /api/v1/puzzles/rush", authenticated(http.HandlerFunc(handler.StartRush)))
	mux.Handle("GET /api/v1/puzzles/rush/{sessionID}", authenticated(http.HandlerFunc(handler.GetRush)))
	mux.Handle("POST /api/v1/puzzles/rush/{sessionID}/answers", authenticated(http.HandlerFunc(handler.SubmitRushAnswer)))
	mux.Handle("GET /api/v1/puzzles/{puzzleID}", authenticated(http.HandlerFunc(handler.Get)))
	mux.Handle("POST /api/v1/puzzles/{puzzleID}/attempts", authenticated(http.HandlerFunc(handler.Submit)))
}
