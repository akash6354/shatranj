package coaches

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
	mux.HandleFunc("GET /api/v1/coaches", handler.List)
	mux.HandleFunc("GET /api/v1/coaches/{coachID}", handler.Get)
	mux.Handle("PUT /api/v1/coaches/me", authenticated(http.HandlerFunc(handler.Upsert)))
	mux.Handle("POST /api/v1/coaches/{coachID}/bookings", authenticated(http.HandlerFunc(handler.Book)))
	mux.Handle("GET /api/v1/coaches/bookings/me", authenticated(http.HandlerFunc(handler.Bookings)))
	mux.Handle("PUT /api/v1/coaches/bookings/{bookingID}", authenticated(http.HandlerFunc(handler.UpdateBooking)))
}
