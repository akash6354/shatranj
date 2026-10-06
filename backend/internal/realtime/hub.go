package realtime

import (
	"log/slog"
	"sync"
	"time"
)

type Hub struct {
	mu     sync.RWMutex
	games  map[string]map[*Client]struct{}
	chats  map[string]map[*ChatClient]struct{}
	users  map[string]int
	logger *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}
	return &Hub{games: make(map[string]map[*Client]struct{}), chats: make(map[string]map[*ChatClient]struct{}), users: make(map[string]int), logger: logger}
}

func (h *Hub) register(client *Client) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.games[client.gameID] == nil {
		h.games[client.gameID] = make(map[*Client]struct{})
	}
	h.games[client.gameID][client] = struct{}{}
	h.users[client.userID]++
	return h.users[client.userID]
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.games[client.gameID]; clients != nil {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)
			if len(clients) == 0 {
				delete(h.games, client.gameID)
			}
			h.users[client.userID]--
			if h.users[client.userID] <= 0 {
				delete(h.users, client.userID)
			}
		}
	}
}

// Publish sends events without blocking game processing. Clients unable to
// drain their event buffers are disconnected and should reload game state.
func (h *Hub) Publish(gameID string, event any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.games[gameID] {
		if !client.enqueue(event) {
			delete(h.games[gameID], client)
			close(client.send)
			h.users[client.userID]--
			if h.users[client.userID] <= 0 {
				delete(h.users, client.userID)
			}
			h.logger.Warn("removed slow WebSocket client", "user_id", client.userID, "game_id", gameID)
		}
	}
	if len(h.games[gameID]) == 0 {
		delete(h.games, gameID)
	}
}

func (h *Hub) Presence(userID string) (bool, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := h.users[userID]
	return count > 0, count
}

// CloseAll closes live sockets during process shutdown.
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, clients := range h.games {
		for client := range clients {
			close(client.send)
			_ = client.conn.Close()
		}
	}
	h.games = make(map[string]map[*Client]struct{})
	for _, clients := range h.chats {
		for client := range clients {
			close(client.send)
			_ = client.conn.Close()
		}
	}
	h.chats = make(map[string]map[*ChatClient]struct{})
	h.users = make(map[string]int)
}

func (h *Hub) registerChat(client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.chats[client.roomID] == nil {
		h.chats[client.roomID] = make(map[*ChatClient]struct{})
	}
	h.chats[client.roomID][client] = struct{}{}
	h.users[client.userID]++
}

func (h *Hub) unregisterChat(client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.chats[client.roomID]; clients != nil {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)
			if len(clients) == 0 {
				delete(h.chats, client.roomID)
			}
			h.users[client.userID]--
			if h.users[client.userID] <= 0 {
				delete(h.users, client.userID)
			}
		}
	}
}

func (h *Hub) PublishRoom(roomID string, event any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.chats[roomID] {
		if !client.enqueue(event) {
			delete(h.chats[roomID], client)
			close(client.send)
			h.users[client.userID]--
			if h.users[client.userID] <= 0 {
				delete(h.users, client.userID)
			}
			h.logger.Warn("removed slow chat WebSocket client", "user_id", client.userID, "room_id", roomID)
		}
	}
	if len(h.chats[roomID]) == 0 {
		delete(h.chats, roomID)
	}
}

func (h *Hub) connected(client *Client, count int) {
	client.enqueue(Event{Type: EventConnected, GameID: client.gameID, UserID: client.userID, CreatedAt: time.Now().UTC()})
	h.Publish(client.gameID, Event{Type: EventPresence, GameID: client.gameID, UserID: client.userID, Online: true, Count: count, CreatedAt: time.Now().UTC()})
}

func (h *Hub) disconnected(client *Client) {
	h.unregister(client)
	online, count := h.Presence(client.userID)
	h.Publish(client.gameID, Event{Type: EventPresence, GameID: client.gameID, UserID: client.userID, Online: online, Count: count, CreatedAt: time.Now().UTC()})
}
