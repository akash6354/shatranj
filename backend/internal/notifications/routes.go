package notifications

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
	mux.Handle("GET /api/v1/notifications", authenticated(http.HandlerFunc(handler.List)))
	mux.Handle("GET /api/v1/notifications/unread-count", authenticated(http.HandlerFunc(handler.UnreadCount)))
	mux.Handle("PUT /api/v1/notifications/{notificationID}/read", authenticated(http.HandlerFunc(handler.MarkRead)))
	mux.Handle("DELETE /api/v1/notifications/{notificationID}/read", authenticated(http.HandlerFunc(handler.MarkUnread)))
}
