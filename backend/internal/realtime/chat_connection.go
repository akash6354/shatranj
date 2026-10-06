package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shatranj/backend/internal/auth"
	"github.com/shatranj/backend/internal/chat"
	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type ChatService interface {
	AuthorizeChat(context.Context, string, string) error
	SendChat(context.Context, string, string, string) (any, error)
}

type ChatClient struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	userID  string
	roomID  string
	logger  *slog.Logger
	service ChatService
	ctx     context.Context
}

func RegisterChatRoute(mux *http.ServeMux, hub *Hub, tokens *auth.TokenManager, allowedOrigins []string, service ChatService) {
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
		ReadBufferSize: 1024, WriteBufferSize: 1024,
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
	mux.Handle("GET /api/v1/chat/rooms/{roomID}/ws", authenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.PrincipalFromContext(r.Context())
		if !ok {
			httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
			return
		}
		roomID := r.PathValue("roomID")
		if err := service.AuthorizeChat(r.Context(), roomID, principal.UserID); err != nil {
			if errors.Is(err, chat.ErrNotFound) {
				httpapi.WriteError(w, http.StatusNotFound, "chat_not_found", "chat room not found")
			} else if errors.Is(err, chat.ErrForbidden) {
				httpapi.WriteError(w, http.StatusForbidden, "forbidden", "chat room membership required")
			} else {
				slog.ErrorContext(r.Context(), "authorize chat WebSocket", "error", err)
				httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
			return
		}
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Debug("upgrade chat WebSocket", "error", err)
			return
		}
		client := &ChatClient{
			hub: hub, conn: connection, send: make(chan []byte, 64),
			userID: principal.UserID, roomID: roomID, logger: slog.Default(), service: service, ctx: r.Context(),
		}
		hub.registerChat(client)
		client.enqueue(Event{Type: EventConnected, Data: map[string]string{"room_id": roomID}, CreatedAt: time.Now().UTC()})
		_, count := hub.Presence(principal.UserID)
		hub.PublishRoom(roomID, Event{
			Type: EventPresenceUpdated, UserID: principal.UserID, Online: true, Count: count, CreatedAt: time.Now().UTC(),
		})
		go client.writePump()
		client.readPump()
		hub.unregisterChat(client)
	})))
}

func (c *ChatClient) readPump() {
	defer func() {
		c.hub.unregisterChat(c)
		online, count := c.hub.Presence(c.userID)
		c.hub.PublishRoom(c.roomID, Event{
			Type: EventPresenceUpdated, UserID: c.userID, Online: online, Count: count, CreatedAt: time.Now().UTC(),
		})
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		_, body, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Debug("chat WebSocket closed", "user_id", c.userID, "error", err)
			}
			return
		}
		var input struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(body, &input); err != nil || input.Type != "chat.send" {
			c.enqueue(Event{Type: "chat.error", Data: map[string]string{"message": "unsupported chat event"}, CreatedAt: time.Now().UTC()})
			continue
		}
		if _, err := c.service.SendChat(c.ctx, c.roomID, c.userID, input.Content); err != nil {
			c.enqueue(Event{Type: "chat.error", Data: map[string]string{"message": "message rejected"}, CreatedAt: time.Now().UTC()})
		}
	}
}

func (c *ChatClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() { ticker.Stop(); _ = c.conn.Close() }()
	for {
		select {
		case body, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, body); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *ChatClient) enqueue(event any) bool {
	body, err := json.Marshal(event)
	if err != nil {
		c.logger.Error("encode chat WebSocket event", "error", err)
		return false
	}
	select {
	case c.send <- body:
		return true
	default:
		return false
	}
}
