package realtime

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shatranj/backend/internal/auth"
	"github.com/shatranj/backend/internal/games"
	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type GameAuthorizer func(context.Context, string, string) error

func RegisterGameRoute(mux *http.ServeMux, hub *Hub, tokens *auth.TokenManager, allowedOrigins []string, authorize GameAuthorizer) {
	verify := func(raw string) (middleware.Principal, error) {
		claims, err := tokens.Verify(raw)
		if err != nil {
			return middleware.Principal{}, err
		}
		return middleware.Principal{UserID: claims.Subject, Email: claims.Email}, nil
	}
	authenticated := middleware.Authenticate(verify)
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			if _, ok := allowed[origin]; ok {
				return true
			}
			return origin == "http://"+r.Host || origin == "https://"+r.Host
		},
	}
	mux.Handle("GET /api/v1/games/{gameID}/ws", authenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.PrincipalFromContext(r.Context())
		if !ok {
			httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
			return
		}
		gameID := r.PathValue("gameID")
		if err := authorize(r.Context(), gameID, principal.UserID); err != nil {
			switch {
			case errors.Is(err, games.ErrNotFound):
				httpapi.WriteError(w, http.StatusNotFound, "game_not_found", "game not found")
			case errors.Is(err, games.ErrForbidden):
				httpapi.WriteError(w, http.StatusForbidden, "forbidden", "you are not a player in this game")
			default:
				slog.ErrorContext(r.Context(), "authorize game WebSocket", "error", err)
				httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
			return
		}
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Debug("upgrade game WebSocket", "error", err)
			return
		}
		client := &Client{
			hub: hub, conn: connection, send: make(chan []byte, 64),
			userID: principal.UserID, gameID: gameID, logger: slog.Default(),
		}
		count := hub.register(client)
		hub.connected(client, count)
		go client.writePump()
		client.readPump()
		hub.disconnected(client)
	})))
}
