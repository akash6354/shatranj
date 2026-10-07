package games

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
	mux.Handle("POST /api/v1/games", authenticated(http.HandlerFunc(handler.Create)))
	mux.Handle("POST /api/v1/games/{gameID}/join", authenticated(http.HandlerFunc(handler.Join)))
	mux.Handle("GET /api/v1/games/{gameID}", authenticated(http.HandlerFunc(handler.Get)))
	mux.Handle("POST /api/v1/games/{gameID}/moves", authenticated(http.HandlerFunc(handler.Move)))
	mux.Handle("POST /api/v1/games/{gameID}/resign", authenticated(http.HandlerFunc(handler.Resign)))
	mux.Handle("POST /api/v1/games/{gameID}/draw-offers", authenticated(http.HandlerFunc(handler.OfferDraw)))
	mux.Handle("POST /api/v1/games/{gameID}/draw-offers/accept", authenticated(http.HandlerFunc(handler.AcceptDraw)))
	mux.Handle("POST /api/v1/games/{gameID}/draw-claims", authenticated(http.HandlerFunc(handler.ClaimDraw)))
}
