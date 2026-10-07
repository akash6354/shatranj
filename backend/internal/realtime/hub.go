package realtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shatranj/backend/internal/cache"
)

type Hub struct {
	mu           sync.RWMutex
	games        map[string]map[*Client]struct{}
	chats        map[string]map[*ChatClient]struct{}
	users        map[string]int
	store        cache.Store
	logger       *slog.Logger
	presence     time.Duration
	nodeID       string
	fanout       cache.PubSubSubscription
	cancelFanout context.CancelFunc
	fanoutDone   chan struct{}
}

const gameEventTopic = "shatranj:game-events:v1"

type gameEventEnvelope struct {
	Origin string          `json:"origin"`
	GameID string          `json:"game_id"`
	Event  json.RawMessage `json:"event"`
}

func NewHub(logger *slog.Logger) *Hub {
	return NewHubWithPresence(logger, nil)
}

func NewHubWithPresence(logger *slog.Logger, store cache.Store) *Hub {
	if logger == nil {
		logger = slog.Default()
	}
	var nodeIDBytes [16]byte
	if _, err := rand.Read(nodeIDBytes[:]); err != nil {
		return &Hub{
			games: make(map[string]map[*Client]struct{}), chats: make(map[string]map[*ChatClient]struct{}),
			users: make(map[string]int), store: store, logger: logger, presence: 2 * time.Minute,
			nodeID: fmt.Sprintf("%d", time.Now().UnixNano()),
		}
	}
	return &Hub{
		games: make(map[string]map[*Client]struct{}), chats: make(map[string]map[*ChatClient]struct{}),
		users: make(map[string]int), store: store, logger: logger, presence: 2 * time.Minute,
		nodeID: hex.EncodeToString(nodeIDBytes[:]),
	}
}

func (h *Hub) EnableGameEventFanout(ctx context.Context) error {
	broker, ok := h.store.(cache.PubSubStore)
	if !ok {
		return nil
	}
	subscription, err := broker.Subscribe(ctx, gameEventTopic)
	if err != nil {
		return fmt.Errorf("subscribe game event fanout: %w", err)
	}
	fanoutCtx, cancel := context.WithCancel(ctx)
	h.fanout, h.cancelFanout = subscription, cancel
	h.fanoutDone = make(chan struct{})
	go func() {
		defer close(h.fanoutDone)
		for {
			select {
			case <-fanoutCtx.Done():
				return
			case message, ok := <-subscription.Messages():
				if !ok {
					return
				}
				var envelope gameEventEnvelope
				if err := json.Unmarshal(message, &envelope); err != nil || envelope.Origin == h.nodeID || envelope.GameID == "" {
					continue
				}
				var event any
				if err := json.Unmarshal(envelope.Event, &event); err != nil {
					h.logger.Warn("decode remote game event", "error", err)
					continue
				}
				h.publishLocal(envelope.GameID, event)
			}
		}
	}()
	return nil
}

func (h *Hub) register(client *Client) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.games[client.gameID] == nil {
		h.games[client.gameID] = make(map[*Client]struct{})
	}
	h.games[client.gameID][client] = struct{}{}
	h.users[client.userID]++
	h.markPresenceLocked(client.userID, h.users[client.userID])
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
			h.markPresenceLocked(client.userID, h.users[client.userID])
		}
	}
}

// Publish sends events without blocking game processing. Clients unable to
// drain their event buffers are disconnected and should reload game state.
func (h *Hub) Publish(gameID string, event any) {
	h.publishLocal(gameID, event)
	if h.fanout == nil {
		return
	}
	encodedEvent, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("encode shared game event", "error", err)
		return
	}
	message, err := json.Marshal(gameEventEnvelope{Origin: h.nodeID, GameID: gameID, Event: encodedEvent})
	if err != nil {
		h.logger.Error("encode shared game event envelope", "error", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := h.store.(cache.PubSubStore).Publish(ctx, gameEventTopic, message); err != nil {
		h.logger.Warn("publish shared game event", "game_id", gameID, "error", err)
	}
}

func (h *Hub) publishLocal(gameID string, event any) {
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
			h.markPresenceLocked(client.userID, h.users[client.userID])
			h.logger.Warn("removed slow WebSocket client", "user_id", client.userID, "game_id", gameID)
		}
	}
	if len(h.games[gameID]) == 0 {
		delete(h.games, gameID)
	}
}

func (h *Hub) Presence(userID string) (bool, int) {
	h.mu.RLock()
	count := h.users[userID]
	store := h.store
	h.mu.RUnlock()
	if count > 0 {
		return true, count
	}
	if store != nil {
		if _, err := store.Get(context.Background(), cache.PresenceKey(userID)); err == nil {
			return true, 0
		} else if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
			h.logger.Debug("read presence fallback failed", "user_id", userID, "error", err)
		}
	}
	return false, 0
}

// CloseAll closes live sockets during process shutdown.
func (h *Hub) CloseAll() {
	if h.cancelFanout != nil {
		h.cancelFanout()
	}
	if h.fanout != nil {
		_ = h.fanout.Close()
	}
	if h.fanoutDone != nil {
		<-h.fanoutDone
	}
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
	h.markPresenceLocked(client.userID, h.users[client.userID])
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
			h.markPresenceLocked(client.userID, h.users[client.userID])
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
			h.markPresenceLocked(client.userID, h.users[client.userID])
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

func (h *Hub) markPresenceLocked(userID string, count int) {
	if h.store == nil || userID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var err error
	if count > 0 {
		err = h.store.Set(ctx, cache.PresenceKey(userID), []byte("online"), h.presence)
	} else {
		err = h.store.Delete(ctx, cache.PresenceKey(userID))
	}
	if err != nil {
		h.logger.Debug("update presence cache", "user_id", userID, "error", err)
	}
}
