package tournaments

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
	mux.Handle("GET /api/v1/tournaments", authenticated(http.HandlerFunc(handler.List)))
	mux.Handle("POST /api/v1/tournaments", authenticated(http.HandlerFunc(handler.Create)))
	mux.Handle("GET /api/v1/tournaments/{tournamentID}", authenticated(http.HandlerFunc(handler.Get)))
	mux.Handle("POST /api/v1/tournaments/{tournamentID}/registration", authenticated(http.HandlerFunc(handler.Register)))
	mux.Handle("DELETE /api/v1/tournaments/{tournamentID}/registration", authenticated(http.HandlerFunc(handler.Withdraw)))
	mux.Handle("POST /api/v1/tournaments/{tournamentID}/rounds", authenticated(http.HandlerFunc(handler.StartRound)))
	mux.Handle("GET /api/v1/tournaments/{tournamentID}/standings", authenticated(http.HandlerFunc(handler.Standings)))
	mux.Handle("POST /api/v1/tournaments/{tournamentID}/pairings/{pairingID}/result", authenticated(http.HandlerFunc(handler.ReportResult)))
	mux.Handle("POST /api/v1/tournaments/{tournamentID}/complete", authenticated(http.HandlerFunc(handler.Complete)))
}
