package chat

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
	mux.Handle("GET /api/v1/chat/rooms", authenticated(http.HandlerFunc(handler.Rooms)))
	mux.Handle("POST /api/v1/chat/rooms", authenticated(http.HandlerFunc(handler.Open)))
	mux.Handle("GET /api/v1/chat/rooms/{roomID}/messages", authenticated(http.HandlerFunc(handler.Messages)))
	mux.Handle("POST /api/v1/chat/rooms/{roomID}/messages", authenticated(http.HandlerFunc(handler.Send)))
	mux.Handle("DELETE /api/v1/chat/rooms/{roomID}/messages/{messageID}", authenticated(http.HandlerFunc(handler.Hide)))
}
